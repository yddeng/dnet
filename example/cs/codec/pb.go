package codec

import (
	"github.com/yddeng/dnet/example/module/protocol"
	"github.com/yddeng/dnet/example/module/protocol/protobuf"
	"github.com/yddeng/dnet/example/pb"
)

func init() {
	protocol.InitProtocol(protobuf.Protobuf{})

	protocol.Register(1, &pb.EchoToS{})
	protocol.Register(2, &pb.EchoToC{})
}
