package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/dtcp"
	"github.com/yddeng/dnet/example/cs/codec"
	"github.com/yddeng/dnet/example/module/message"
	"github.com/yddeng/dnet/example/pb"
	"github.com/yddeng/dutil/buffer"
	"time"
)

func main() {
	addr := "localhost:1234"
	session, err := dtcp.DialTCP("tcp", addr, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("conn ok,remote:%s\n", session.RemoteAddr())

	session.SetCodec(codec.NewCodec())
	session.SetCloseCallBack(func(session dnet.Session, reason string) {
		fmt.Println("onClose", reason)
	})
	_ = session.Start(func(data interface{}, err2 error) {
		//fmt.Println("data", data, "err", err)
		if err2 != nil {
			session.Close(err2.Error())
		} else {
			fmt.Println("read ", data.(*message.Message).GetData())
		}
	})

	fmt.Println(session.Send(message.NewMessage(0, &pb.EchoToS{Msg: proto.String("hi server")})))
	fmt.Println(session.Send(message.NewMessage(0, &pb.EchoToS{Msg: proto.String("hi server")})))
	time.Sleep(5 * time.Second)

	bt, _ := proto.Marshal(&pb.EchoToS{Msg: proto.String("hi server")})
	msgLen := len(bt) + 2 + 2 + 2
	buff := buffer.NewBufferWithCap(msgLen)
	//msgLen+cmd+msgID
	//写入data长度
	buff.WriteUint16BE(uint16(len(bt)))
	//写入cmd
	buff.WriteUint16BE(0)
	//msgID
	buff.WriteUint16BE(1)
	//data数据
	buff.WriteBytes(bt)
	fmt.Println(session.SendBytes(buff.Bytes()))

	select {}

}
