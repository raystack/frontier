package echo

import (
	"context"
	"fmt"
	"github.com/odpf/shield/gen/go/protos"
)

type SimpleServer struct {
	protos.UnimplementedYourServiceServer
}

func (s *SimpleServer) Echo(ctx context.Context, msg *protos.StringMessage) (*protos.StringMessage, error) {
	fmt.Printf("Received message: %s", msg.Payload.Validate())
	return &protos.StringMessage{Id: "10"}, nil
}

//func (s
