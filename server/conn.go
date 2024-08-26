package server

import "github.com/sandertv/gophertunnel/minecraft"

type Conn struct {
	ServerToConnect *RemoteConfig
	*minecraft.Conn
}
