package pb

import (
	"github.com/tagDong/dnet/codec/protobuf"
	"github.com/tagDong/dnet/module/protocol"
)

func init() {
	protocol.InitProtocol(protobuf.Protobuf{})

	protocol.RegisterIDMsg(1, &EchoToS{})
	protocol.RegisterIDMsg(2, &EchoToC{})
}
