package codec

import (
	"github.com/yddeng/dnet/example/module/codec/protobuf"
	"github.com/yddeng/dnet/example/module/protocol"
	"github.com/yddeng/dnet/example/pb"
)

func init() {
	protocol.InitProtocol(protobuf.Protobuf{})

	protocol.RegisterIDMsg(1, &pb.EchoToS{})
	protocol.RegisterIDMsg(2, &pb.EchoToC{})
}
