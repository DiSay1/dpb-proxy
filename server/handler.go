package server

import (
	"encoding/json"
	"log"
	"sync"
	"time"

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

type userSeclet struct {
	unixTime int64
	conf     *RemoteConfig
}

var userServers = map[string]*userSeclet{}

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

	userSelected := &userSeclet{}
	res, ok := userServers[playerIdentity.XUID]

	if !ok {
		if err := conn.StartGame(minecraft.GameData{}); err != nil {
			log.Println("not start game, err:", err)

			return
		}

		log.Printf("%v joined!\n", conn.IdentityData().DisplayName)

		serverInfo := s.handleSelectServer(conn.Conn)
		if serverInfo == nil {
			return
		}

		userServers[playerIdentity.XUID] = &userSeclet{
			unixTime: time.Now().Unix(),
			conf:     serverInfo,
		}

		conn.WritePacket(&packet.Transfer{
			Address: "51.68.143.191",
			Port:    uint16(8080),
		})
		return
	} else {
		if time.Now().Unix()-res.unixTime > 108000 {
			if err := conn.StartGame(minecraft.GameData{}); err != nil {
				log.Println("not start game, err:", err)

				return
			}

			log.Printf("%v joined!\n", conn.IdentityData().DisplayName)

			serverInfo := s.handleSelectServer(conn.Conn)
			if serverInfo == nil {
				return
			}

			userServers[playerIdentity.XUID] = &userSeclet{
				unixTime: time.Now().Unix(),
				conf:     serverInfo,
			}

			conn.WritePacket(&packet.Transfer{
				Address: s.cfg.Listener.PublicAddress,
				Port:    s.cfg.Listener.PublicPort,
			})
			return
		} else {
			log.Printf("%v selected server!\n", playerIdentity.DisplayName)
			userSelected = userServers[playerIdentity.XUID]
		}
	}

	serverConn, err := minecraft.Dialer{
		IdentityData:        conn.IdentityData(),
		ClientData:          conn.ClientData(),
		KeepXBLIdentityData: true,
	}.Dial("raknet", userSelected.conf.Address)
	if err != nil {
		panic(err)
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

	connected := true
	go func() {
		if err := conn.StartGame(serverConn.GameData()); err != nil {
			connected = false
		}
		g.Done()
	}()

	go func() {
		if err := serverConn.DoSpawn(); err != nil {
			connected = false
		}
		g.Done()
	}()

	g.Wait()

	if connected {
		go s.toServer(conn.Conn, serverConn)
		s.toClient(conn.Conn, serverConn)
	} else {
		return
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
			pIO := s.proto.NewReader(&pk.Payload, 0, false)
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
