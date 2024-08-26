package server

import (
	"errors"
	"log"

	"github.com/sandertv/gophertunnel/minecraft"
	"proj.dichay.tech/dpb-proxy/decoder"
)

func (s *Server) toClient(clientConn *minecraft.Conn, serverConn *minecraft.Conn) {
	defer func() {
		playerInfo := clientConn.IdentityData()
		err := clientConn.Close()
		if err != nil {
			log.Println("not close connection, err", err)
		}

		err = serverConn.Close()
		if err != nil {
			log.Println("not close connection, err", err)
		}

		delete(s.connections, playerInfo.XUID)
	}()

	for {
		data, err := serverConn.ReadBytes()
		if err != nil {
			var disc minecraft.DisconnectError

			if ok := errors.As(err, &disc); ok {
				_ = s.listener.Disconnect(clientConn, disc.Error())
			}

			return
		}

		pk, err := decoder.ParseData(data)
		if err != nil {
			log.Println("not deocde packet, err:", err)
		} else {
			for _, module := range s.modules {
				ok, err := module.ToClient(clientConn, serverConn, pk)
				if err != nil {
					log.Println(err, ok)
				} else {
					if ok {
						continue
					}
				}
			}
		}

		if _, err := clientConn.Write(data); err != nil {
			return
		}
	}
}
