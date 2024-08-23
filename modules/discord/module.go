package discord

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"proj.dichay.tech/dpb-proxy/decoder"
)

type DiscordModule struct {
	//
}

func (dm *DiscordModule) Init() error {
	//
	return nil
}

func (dm *DiscordModule) ToServer(conn *minecraft.Conn, pk *decoder.PacketData) error {
	return nil
}

func (dm *DiscordModule) ToClient(conn *minecraft.Conn, pk *decoder.PacketData) error {
	return nil
}
