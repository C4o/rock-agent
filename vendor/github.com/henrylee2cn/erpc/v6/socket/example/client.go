package main

import (
	"log"
	"net"

	"github.com/henrylee2cn/erpc/v6/codec"
	"github.com/henrylee2cn/erpc/v6/socket"
	"github.com/henrylee2cn/erpc/v6/socket/example/pb"
)

//go:generate go build $GOFILE

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8000")
	if err != nil {
		log.Fatalf("[CLI] dial err: %v", err)
	}
	s := socket.GetSocket(conn)
	defer s.Close()
	var message = socket.GetMessage()
	defer socket.PutMessage(message)
	for i := int32(0); i < 1; i++ {
		// write request
		message.Reset()
		message.SetMtype(0)
		message.SetBodyCodec(codec.ID_JSON)
		message.SetSeq(i)
		message.SetServiceMethod("/a/b")
		message.SetBody(&pb.PbTest{A: 10, B: 2})
		err = s.WriteMessage(message)
		if err != nil {
			log.Printf("[CLI] write request err: %v", err)
			continue
		}
		log.Printf("[CLI] write request: %s", message.String())

		// read response
		message.Reset(socket.WithNewBody(
			func(header socket.Header) interface{} {
				return new(pb.PbTest)
			}),
		)
		err = s.ReadMessage(message)
		if err != nil {
			log.Printf("[CLI] read response err: %v", err)
		} else {
			log.Printf("[CLI] read response: %s", message.String())
		}
	}
	// select {}
}
