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

		c, ok := conn.(*minecraft.Conn)
		if !ok {
			c.Close()
			continue
		}

		playerIdentity := c.IdentityData()
		_, ok = s.connections[playerIdentity.XUID]
		if ok {
			err := c.Close()
			if err != nil {
				log.Println("not close connection, err", err)
			}
			continue
		}
		s.connections[playerIdentity.XUID] = &Conn{c}

		go s.handleConn(c)
	}
}
