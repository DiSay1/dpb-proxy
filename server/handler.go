package server

import (
	"encoding/json"
	"log"
	"time"

	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"proj.dichay.tech/dpb-proxy/decoder"
)

type ServerSelect struct {
	Servers form.Dropdown
}

func (f ServerSelect) Submit(submitter form.Submitter) {
}

func (s *Server) handleConn(conn *minecraft.Conn) {
	log.Printf("%v joined!\n", conn.IdentityData().DisplayName)

	if err := conn.StartGame(minecraft.GameData{}); err != nil {
		log.Println("not start game, err:", err)

		return
	}

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
		return
	}

	serverInfo := s.choiceServer(conn)

	serverConn, err := minecraft.Dialer{
		IdentityData:        conn.IdentityData(),
		ClientData:          conn.ClientData(),
		KeepXBLIdentityData: true,
	}.Dial("raknet", serverInfo.Address)
	if err != nil {
		panic(err)
	}

	defer func() {
		playerInfo := conn.IdentityData()
		err := conn.Close()
		if err != nil {
			log.Println("not close connection, err", err)
		}

		err = serverConn.Close()
		if err != nil {
			log.Println("not close connection, err", err)
		}

		delete(s.connections, playerInfo.XUID)
	}()

	d := serverConn.GameData()

	conn.WritePacket(&packet.ChangeDimension{
		Dimension:       d.Dimension,
		Position:        d.PlayerPosition,
		LoadingScreenID: protocol.Option(uint32(time.Now().Unix())),
	})
	conn.WritePacket(&packet.StopSound{StopAll: true})
	conn.WritePacket(&packet.PlayStatus{Status: packet.PlayStatusPlayerSpawn})

	conn.WritePacket(&packet.PlayerAction{
		EntityRuntimeID: d.EntityRuntimeID,
		ActionType:      protocol.PlayerActionDimensionChangeDone,
	})

	conn.WritePacket(&packet.MovePlayer{
		EntityRuntimeID: d.EntityRuntimeID,

		Position: d.PlayerPosition,
		Pitch:    d.Pitch,
		Yaw:      d.Yaw,
		HeadYaw:  d.Yaw,
		Mode:     packet.MoveModeTeleport,
	})

	conn.WritePacket(&packet.MoveActorAbsolute{
		EntityRuntimeID: d.EntityRuntimeID,
		Position:        d.PlayerPosition,
		Rotation:        mgl32.Vec3{d.Pitch, d.Yaw, d.Yaw},
		Flags:           packet.MoveFlagTeleport,
	})

	go s.toServer(conn, serverConn)
	s.toClient(conn, serverConn)
}

func (s *Server) choiceServer(conn *minecraft.Conn) *RemoteConfig {
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
