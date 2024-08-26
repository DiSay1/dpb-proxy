package decoder

import (
	"bytes"
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type PacketData struct {
	H       *packet.Header
	Payload bytes.Buffer
}

func ParseData(data []byte) (*PacketData, error) {
	buf := bytes.NewBuffer(data)
	header := &packet.Header{}
	if err := header.Read(buf); err != nil {
		// We don't return this as an error as it's not in the hand of the user to control this. Instead,
		// we return to reading a new packet.

		return nil, fmt.Errorf("read packet header: %w", err)
	}

	return &PacketData{header, *buf}, nil
}
