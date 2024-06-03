package eiscp

import (
	"bytes"
	"encoding/binary"
	"strings"
)

// The eISCP packet wraps ISCP message for communication over Ethernet
type EISCPPacket struct {
	Magic      [4]byte
	HeaderSize uint32
	DataSize   uint32
	Version    byte
	Reserved   [3]byte
	Data       []byte
}

func (p *EISCPPacket) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, p.Magic)
	binary.Write(buf, binary.BigEndian, p.HeaderSize)
	binary.Write(buf, binary.BigEndian, p.DataSize)
	binary.Write(buf, binary.BigEndian, p.Version)
	binary.Write(buf, binary.BigEndian, p.Reserved)
	buf.Write(p.Data)
	return buf.Bytes()
}

func NewEISCPPacket(iscpMessage string) *EISCPPacket {
	iscpMessage = "!1" + iscpMessage + "\r"
	iscpMessageBytes := []byte(iscpMessage)
	return &EISCPPacket{
		Magic:      [4]byte{'I', 'S', 'C', 'P'},
		HeaderSize: 16,
		DataSize:   uint32(len(iscpMessageBytes)),
		Version:    0x01,
		Reserved:   [3]byte{0x00, 0x00, 0x00},
		Data:       iscpMessageBytes,
	}
}

func UnpackEISCPMessage(packet string) string {
	if len(packet) < 16 {
		return packet
	}
	header := packet[:16]
	dataSize := binary.BigEndian.Uint32([]byte(header[8:12]))
	data := packet[16 : 16+dataSize]

	// Remove the ISCP start character '!1' and the trailing '\r'
	message := strings.TrimSpace(string(data[2 : len(data)-3]))
	return message
}
