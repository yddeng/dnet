package pb

import "github.com/tagDong/dnet/codec/protobuf"

func init() {
	protobuf.Register(1, &EchoToS{})
	protobuf.Register(2, &EchoToC{})
}
