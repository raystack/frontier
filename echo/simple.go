package protos

import (
	"context"
	"fmt"
)

type SimpleServer struct {
	UnimplementedYourServiceServer
}

func (s *SimpleServer) Echo(ctx context.Context, msg *StringMessage) (*StringMessage, error) {
	fmt.Printf("Received message: %s", msg.Payload.Validate())
	return &StringMessage{Id: "10"}, nil
}

func (s *SimpleServer) mustEmbedUnimplementedYourServiceServer() {

}
