package server

import (
	"encoding/json"
	"fmt"
	"log"

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

func (s *Server) handleConn(conn *minecraft.Conn) {
	if err := conn.StartGame(minecraft.GameData{}); err != nil {
		log.Println("not start game, err:", err)

		return
	}

	serversName := []string{}
	for _, server := range s.cfg.Servers {
		serversName = append(serversName, server.Name)
	}
	serverForm := form.New(ServerSelect{
		Servers: form.NewDropdown("СЕРВЕРА", serversName, 0),
	}, "ВЫБЕРИТЕ СЕРВЕР")

	f := form.Form(serverForm)
	data, _ := json.Marshal(f)

	conn.WritePacket(&packet.ModalFormRequest{
		FormID:   0,
		FormData: data,
	})

	for {
		data, err := conn.ReadBytes()
		if err != nil {
			log.Println("not read packet, err:", err)
			continue
		}

		pk, err := decoder.ParseData(data)
		if err != nil {
			log.Println("not parse packet, err:", err)
			continue
		}

		switch pk.H.PacketID {
		case packet.IDModalFormResponse:
			form := packet.ModalFormResponse{}
			pIO := s.proto.NewReader(pk.Payload, 0, false)
			form.Marshal(pIO)

			fmt.Println(form)
			d, _ := form.ResponseData.Value()
			fmt.Printf("string(d): %v\n", string(d))
		}
	}
}
