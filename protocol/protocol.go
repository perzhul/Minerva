package protocol

type Packet struct {
	Length   uint64 // Length of Packet ID + Data
	PacketID uint64 // Corresponds to protocol_id from the server's packet report
}

type StatusResponsePacket struct {
	Packet
	JSONResponse string // prefixed by its length as a VarInt(3-byte max)
}

// packetLength Varint, packetID Varint
func NewStatusResponse(packetLength uint64, packetID uint64, JSONResponse string) StatusResponsePacket {
	return StatusResponsePacket{
		Packet: Packet{
			Length:   packetLength,
			PacketID: packetID,
		},
		JSONResponse: JSONResponse,
	}
}
