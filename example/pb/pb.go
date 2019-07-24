package pb

import (
	"github.com/tagDong/dnet/codec/protobuf"
	"github.com/tagDong/dnet/module/protocol"
)

var PbMate *protocol.Protocol

func init() {
	PbMate = protocol.NewProtocol(protobuf.Protobuf{})
	PbMate.RegisterIDMsg(1, &EchoToS{})
	PbMate.RegisterIDMsg(2, &EchoToC{})
}
