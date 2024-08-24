package logger

import (
	"log"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"proj.dichay.tech/dpb-proxy/decoder"
	"proj.dichay.tech/dpb-proxy/server"
)

type LoggerModule struct {
	server *server.Server
	proto  minecraft.Protocol
}

func (dm *LoggerModule) Init(srv any) error {
	server := srv.(*server.Server)

	dm.proto = server.GetProto()
	dm.server = server

	return nil
}

func (dm *LoggerModule) ToServer(conn *minecraft.Conn, serverConn *minecraft.Conn, pk *decoder.PacketData) (bool, error) {
	switch pk.H.PacketID {
	case packet.IDText:
		text := packet.Text{}
		pIO := dm.proto.NewReader(pk.Payload, 0, false)
		text.Marshal(pIO)

		log.Printf("%v: %v\n", text.SourceName, text.Message)
	}
	return false, nil
}

func (dm *LoggerModule) ToClient(conn *minecraft.Conn, serverConn *minecraft.Conn, pk *decoder.PacketData) (bool, error) {
	return false, nil
}
