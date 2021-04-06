package codec

import (
	"github.com/yddeng/dnet/examples/module/protocol"
	"github.com/yddeng/dnet/examples/module/protocol/protobuf"
	"github.com/yddeng/dnet/examples/pb"
)

func init() {
	protocol.InitProtocol(protobuf.Protobuf{})

	protocol.Register(1, &pb.EchoToS{})
	protocol.Register(2, &pb.EchoToC{})
}
