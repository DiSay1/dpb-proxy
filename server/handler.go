package server

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"proj.dichay.tech/dpb-proxy/decoder"
)

type ServerSelect struct {
	Servers form.Dropdown
}

func (f ServerSelect) Submit(submitter form.Submitter) {
}

func (s *Server) handleConn(c *minecraft.Conn) {
	playerIdentity := c.IdentityData()
	defer func() {
		err := c.Close()
		if err != nil {
			log.Println("not close connection, err", err)
		}

		delete(s.connections, playerIdentity.XUID)
	}()

	conn, ok := s.connections[playerIdentity.XUID]
	if !ok {
		return
	}

	if err := conn.StartGame(minecraft.GameData{}); err != nil {
		log.Println("not start game, err:", err)

		return
	}

	serverInfo := &RemoteConfig{}
	if conn.ServerToConnect == nil {
		if err := conn.StartGame(minecraft.GameData{}); err != nil {
			log.Println("not start game, err:", err)

			return
		}

		log.Printf("%v joined!\n", conn.IdentityData().DisplayName)

		serverInfo = s.handleSelectServer(conn.Conn)
		if serverInfo == nil {
			return
		}

		conn.ServerToConnect = serverInfo

		addr := strings.Split(s.cfg.Listener.Address, ":")
		port, _ := strconv.Atoi(addr[1])

		conn.WritePacket(&packet.Transfer{
			Address: addr[0],
			Port:    uint16(port),
		})
	} else {
		serverInfo = conn.ServerToConnect
	}

	serverConn, err := minecraft.Dialer{
		IdentityData:        conn.IdentityData(),
		ClientData:          conn.ClientData(),
		KeepXBLIdentityData: true,
	}.Dial("raknet", serverInfo.Address)
	if err != nil {
		panic(err)
	}

	if err := serverConn.DoSpawn(); err != nil {
		log.Println("not spawn user, err:", err)
		return
	}

	defer func() {
		err = serverConn.Close()
		if err != nil {
			log.Println("not close connection, err", err)
		}

		delete(s.connections, playerIdentity.XUID)
	}()

	var g sync.WaitGroup
	g.Add(2)

	connected := false
	go func() {
		if err := conn.StartGame(serverConn.GameData()); err != nil {
			connected = true
		}
		g.Done()
	}()

	go func() {
		if err := serverConn.DoSpawn(); err != nil {
			connected = true
		}
		g.Done()
	}()

	g.Wait()

	if connected {
		go s.toServer(conn.Conn, serverConn)
		s.toClient(conn.Conn, serverConn)
	}
}

func (s *Server) handleSelectServer(conn *minecraft.Conn) *RemoteConfig {
	serversName := make([]string, len(s.cfg.Servers))
	for i, server := range s.cfg.Servers {
		serversName[i] = server.Name
	}
	serverForm := form.New(ServerSelect{
		Servers: form.NewDropdown("СЕРВЕРА", serversName, 0),
	}, "ВЫБЕРИТЕ СЕРВЕР")

	f := form.Form(serverForm)
	data, _ := json.Marshal(f)

	err := conn.WritePacket(&packet.ModalFormRequest{
		FormID:   0,
		FormData: data,
	})
	if err != nil {
		log.Println("not write packet, err:", err)
		return nil
	}

	serverSelected := false

	server := RemoteConfig{}
	for !serverSelected {
		data, err := conn.ReadBytes()
		if err != nil {
			log.Println("not read packet, err:", err)
			return nil
		}

		pk, err := decoder.ParseData(data)
		if err != nil {
			log.Println("not parse packet, err:", err)
			continue
		}

		switch pk.H.PacketID {
		case packet.IDModalFormResponse:
			formResponse := packet.ModalFormResponse{}
			pIO := s.proto.NewReader(pk.Payload, 0, false)
			formResponse.Marshal(pIO)

			data, ok := formResponse.ResponseData.Value()
			if !ok {
				return nil
			}

			res := []int{}
			err := json.Unmarshal(data, &res)
			if err != nil {
				log.Println("not unmarshal form, err:", err)
			}

			server = s.cfg.Servers[res[0]]
			serverSelected = true
		}
	}

	return &server
}
