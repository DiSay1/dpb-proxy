package server

import (
	"log"

	"github.com/sandertv/gophertunnel/minecraft"
)

func (s *Server) listen() {
	listener := s.listener

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("not accept connection, err", err)
			continue
		}

		minecraftConn, ok := conn.(*minecraft.Conn)
		if !ok {
			minecraftConn.Close()
			continue
		}

		playerIdentity := minecraftConn.IdentityData()
		serverConn, ok := s.connections[playerIdentity.XUID]
		if ok {
			if serverConn.ServerToConnect != nil {
				go s.handleConn(minecraftConn)
			}
			err := minecraftConn.Close()
			if err != nil {
				log.Println("not close connection, err", err)
			}
			continue
		}
		s.connections[playerIdentity.XUID] = &Conn{
			ServerToConnect: nil,
			Conn:            minecraftConn,
		}

		go s.handleConn(minecraftConn)
	}
}
