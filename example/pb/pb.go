package pb

import (
	"github.com/yddeng/dnet/example/module/codec/protobuf"
	"github.com/yddeng/dnet/example/module/protocol"
)

func init() {
	protocol.InitProtocol(protobuf.Protobuf{})

	protocol.RegisterIDMsg(1, &EchoToS{})
	protocol.RegisterIDMsg(2, &EchoToC{})
}
