package daio

import (
	"fmt"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dutil/buffer"
	"io"
	"reflect"
	"testing"
	"time"
)

const (
	lenSize  = 2       // 消息长度（消息体的长度）
	headSize = lenSize // 消息头长度
	buffSize = 65535   // 缓存容量(与lenSize有关，2字节最大65535）
)

type Codec struct {
	readBuf *buffer.Buffer
	dataLen uint16
}

func newCodec() *Codec {
	return &Codec{
		readBuf: buffer.NewBufferWithCap(buffSize),
		dataLen: 0,
	}
}

//解码
func (decoder *Codec) Decode(reader io.Reader) (interface{}, error) {

	for {
		msg, err := decoder.unPack()

		if msg != nil {
			return msg, nil
		} else if err == nil {
			n, err1 := decoder.readBuf.ReadFrom(reader)
			if err1 != nil {
				return nil, err1
			}
			if n <= 0 {
				return nil, nil
			}
		} else {
			return nil, err
		}
	}

}

func (decoder *Codec) unPack() ([]byte, error) {

	if decoder.dataLen == 0 {
		if decoder.readBuf.Len() < headSize {
			return nil, nil
		}

		decoder.dataLen, _ = decoder.readBuf.ReadUint16BE()
	}

	if decoder.readBuf.Len() < int(decoder.dataLen) {
		return nil, nil
	}

	data, _ := decoder.readBuf.ReadBytes(int(decoder.dataLen))

	//将消息长度置为0，用于下一次验证
	decoder.dataLen = 0
	return data, nil
}

//编码
func (encoder *Codec) Encode(o interface{}) ([]byte, error) {
	data, ok := o.([]byte)
	if !ok {
		return nil, fmt.Errorf("encode interface{} is %s, need type []byte", reflect.TypeOf(o))
	}

	dataLen := len(data)
	if dataLen > buffSize-headSize {
		return nil, fmt.Errorf("encode dataLen is too large,len: %d", dataLen)
	}

	msgLen := dataLen + headSize
	buff := buffer.NewBufferWithCap(msgLen)

	//写入data长度
	buff.WriteUint16BE(uint16(dataLen))
	//data数据
	buff.WriteBytes(data)

	return buff.Bytes(), nil
}

func TestEcho(t *testing.T) {
	s := NewService(1)

	ln, err := NewListener("tcp4", "127.0.0.1:9654", s)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	err = ln.Listen(func(session dnet.Session) {
		fmt.Println("new client", session.RemoteAddr().String())
		// 超时时间
		session.SetCodec(newCodec())
		session.SetCloseCallBack(func(reason string) {
			fmt.Println("111 close:", reason)
		})
		errr := session.Start(func(data interface{}, err error) {
			//fmt.Println("data", data, "err", err)
			if err != nil {
				fmt.Println(99, err)
				session.Close(err.Error())
			} else {
				fmt.Println("111 read", data.([]byte))
			}
		})
		if errr != nil {
			fmt.Printf("%s\n", err)
		}

		time.Sleep(time.Second)
		fmt.Println("111 send", session.Send([]byte{3, 2, 1}))

		session.Close("111 call close")
	})
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	fmt.Println("init client start")

	w, err := NewEventLoop()
	if err != nil {
		fmt.Println(err)
		return
	}
	go func() {
		err = w.Run()
		if err != nil {
			fmt.Println(err)
			return
		}
	}()

	cli, err := Dial("127.0.0.1:9654", w)
	if err != nil {
		fmt.Println(err)
		return
	}

	cli.SetCodec(newCodec())
	cli.SetCloseCallBack(func(reason string) {
		fmt.Println("222 close:", reason)
	})
	err = cli.Start(func(message interface{}, err error) {
		//fmt.Println("data", data, "err", err)
		if err != nil {
			fmt.Println(88, err)
			cli.Close(err.Error())
		} else {
			fmt.Println("222 read", message.([]byte))
		}
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("222 send", cli.Send([]byte{1, 2, 3}))
	time.Sleep(time.Second * 2)
	fmt.Println("222 send", cli.Send([]byte{1, 2, 3}))

	time.Sleep(5 * time.Second)
	cli.Close("222 call close")

	select {}
}
