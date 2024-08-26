package discord

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"proj.dichay.tech/dpb-proxy/decoder"
	"proj.dichay.tech/dpb-proxy/server"
)

type Message struct {
	Author string `json:"author"`
	Text   string `json:"text"`
}

type DiscordModule struct {
	lastToServerID   int
	toServerMessages map[int]Message
	toClientMessages map[int]Message

	server *server.Server
	proto  minecraft.Protocol
}

func (dm *DiscordModule) Init(srv any) error {
	server := srv.(*server.Server)

	dm.proto = server.GetProto()
	dm.server = server

	dm.toServerMessages = map[int]Message{}
	dm.toClientMessages = map[int]Message{}

	http.HandleFunc("/api/getMessage", dm.getMessage)
	http.HandleFunc("/api/sendMessage", dm.sendMessage)

	go func() {
		http.ListenAndServe(":8080", nil)
	}()

	return nil
}

func (dm *DiscordModule) ToServer(conn *minecraft.Conn, serverConn *minecraft.Conn, pk *decoder.PacketData) (bool, error) {
	switch pk.H.PacketID {
	case packet.IDText:
		text := packet.Text{}
		pIO := dm.proto.NewReader(&pk.Payload, 0, false)
		text.Marshal(pIO)

		dm.lastToServerID = dm.lastToServerID + 1
		dm.toServerMessages[dm.lastToServerID] = Message{
			Author: text.SourceName,
			Text:   text.Message,
		}
	}
	return false, nil
}

func (dm *DiscordModule) ToClient(conn *minecraft.Conn, serverConn *minecraft.Conn, pk *decoder.PacketData) (bool, error) {
	return false, nil
}

func (dm *DiscordModule) sendMessage(rw http.ResponseWriter, req *http.Request) {
	//
}

func (dm *DiscordModule) getMessage(rw http.ResponseWriter, req *http.Request) {
	if len(dm.toServerMessages) > 0 {
		msgs := []Message{}
		for i := dm.lastToServerID; i > 0; i-- {
			msg, ok := dm.toServerMessages[i]
			if !ok {
				break
			}

			msgs = append(msgs, msg)
		}

		data, err := json.Marshal(msgs)
		if err != nil {
			log.Println(err)
			return
		}

		_, err = rw.Write(data)
		if err != nil {
			log.Println(err)
			return
		}
	}
}
