package server

import (
	"errors"
	"log"

	"github.com/sandertv/gophertunnel/minecraft"
	"proj.dichay.tech/dpb-proxy/decoder"
)

func (s *Server) toServer(clientConn *minecraft.Conn, serverConn *minecraft.Conn) {
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
		data, err := clientConn.ReadBytes()
		if err != nil {
			return
		}

		pk, err := decoder.ParseData(data)
		if err != nil {
			log.Println("not deocde packet, err:", err)
		} else {
			for _, module := range s.modules {
				ok, err := module.ToServer(clientConn, serverConn, pk)
				if err != nil {
					log.Println(err, ok)
				} else {
					if ok {
						continue
					}
				}
			}
		}

		if _, err := serverConn.Write(data); err != nil {
			var disc minecraft.DisconnectError

			if ok := errors.As(err, &disc); ok {
				_ = s.listener.Disconnect(clientConn, disc.Error())
			} else {
				log.Println("not write data to server, err:", err)
			}

			return
		}
	}
}
