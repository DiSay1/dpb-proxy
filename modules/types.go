package modules

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"proj.dichay.tech/dpb-proxy/decoder"
)

type Module interface {
	ToServer(clientConn *minecraft.Conn, serverConn *minecraft.Conn, pk *decoder.PacketData) (bool, error)
	ToClient(clientConn *minecraft.Conn, serverConn *minecraft.Conn, pk *decoder.PacketData) (bool, error)
	Init(server any) error
}
