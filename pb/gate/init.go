package gate

import (
	"io"

	"github.com/golang/protobuf/proto"
)

func DecodePacket(packet []byte) (int, proto.Message, error) {
	return 0,nil,nil
}
func EncodeToPacket(w io.Writer, seq int, pMsg proto.Message) error {
	return nil
}
