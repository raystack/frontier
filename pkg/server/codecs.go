package server

import (
	"errors"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// conectCodec implements https://pkg.go.dev/github.com/bufbuild/connect-go#Codec
// by overriding the default implementation with extra option: protojson.MarshalOptions{UseProtoNames: true}
type conectCodec struct {
}

func (c conectCodec) Name() string {
	return jsonCodec
}

func (c conectCodec) Marshal(message any) ([]byte, error) {
	protoMessage, ok := message.(proto.Message)
	if !ok {
		return nil, ErrNotProto
	}
	return protojson.MarshalOptions{UseProtoNames: true}.Marshal(protoMessage)
}

func (c conectCodec) Unmarshal(binary []byte, message any) error {
	protoMessage, ok := message.(proto.Message)
	if !ok {
		return ErrNotProto
	}
	if len(binary) == 0 {
		return errors.New("zero-length payload is not a valid JSON object")
	}
	// Discard unknown fields so clients and servers aren't forced to always use
	// exactly the same version of the schema.
	options := protojson.UnmarshalOptions{DiscardUnknown: true}
	err := options.Unmarshal(binary, protoMessage)
	if err != nil {
		return ErrNotProto
	}
	return nil
}
