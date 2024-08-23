package server

import (
	"log"

	"github.com/sandertv/gophertunnel/minecraft"
	"proj.dichay.tech/dpb-proxy/modules"
)

type Server struct {
	listener     *minecraft.Listener
	listenConfig *minecraft.ListenConfig
	connections  map[string]*Conn

	modules []modules.Module
	proto   minecraft.Protocol

	cfg *ServerConfig
}

func New(cfg *ServerConfig, modules []modules.Module) *Server {
	listenConfig := &minecraft.ListenConfig{
		AuthenticationDisabled: cfg.Game.AuthenticationDisabled,
		MaximumPlayers:         cfg.Game.MaximumPlayers,
	}

	srv := &Server{
		cfg:          cfg,
		listenConfig: listenConfig,
		modules:      modules,
		proto:        minecraft.DefaultProtocol,

		connections: map[string]*Conn{},
	}

	return srv
}

func (s *Server) Listen() {
	listener, err := s.listenConfig.Listen(`raknet`, s.cfg.Listener.Address)
	if err != nil {
		log.Fatalln("not listen minecraft server, err:", err)
	}

	for _, module := range s.modules {
		err := module.Init()
		if err != nil {
			log.Fatalln("not init module, err:", err)
		}
	}

	s.listener = listener
	s.listen()
}
