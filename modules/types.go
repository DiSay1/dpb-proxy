package modules

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"proj.dichay.tech/dpb-proxy/decoder"
)

type Module interface {
	ToServer(conn *minecraft.Conn, pk *decoder.PacketData) error
	ToClient(conn *minecraft.Conn, pk *decoder.PacketData) error
	Init() error
}
