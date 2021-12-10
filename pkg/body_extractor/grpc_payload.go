package body_extractor

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/jhump/protoreflect/dynamic"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCPayloadCompressionFormat tells by reading the first byte compressed or not
type GRPCPayloadCompressionFormat uint8

const (
	compressionNone GRPCPayloadCompressionFormat = 0
	compressionMade GRPCPayloadCompressionFormat = 1
	maxInt                                       = int(^uint(0) >> 1)
)

type GRPCPayloadHandler struct{}

func (b GRPCPayloadHandler) Extract(body *io.ReadCloser, protoIndex int) (string, error) {
	reqBody, err := ioutil.ReadAll(*body)
	if err != nil {
		return "", err
	}
	defer (*body).Close()
	// repopulate body
	*body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

	return b.extractFromRequest(reqBody, protoIndex)
}

func (b GRPCPayloadHandler) extractFromRequest(body []byte, protoIndex int) (string, error) {
	reqParser := grpcRequestParser{
		r:      bytes.NewBuffer(body),
		header: [5]byte{},
	}
	pf, msg, err := reqParser.Parse()
	if err != nil {
		return "", err
	}
	if pf == compressionMade {
		// unsupported for now
		return "", errors.New("compressed message, unsupported grpc feature")
	}

	return fieldFromProtoMessage(msg, protoIndex)
}

// grpcRequestParser reads complete gRPC messages from the underlying reader
type grpcRequestParser struct {
	// r is the underlying reader
	r io.Reader

	// The header of a gRPC message. Find more detail at
	// https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md
	header [5]byte
}

// Parse reads a complete gRPC message from the stream.
//
// It returns the message and its payload (compression/encoding)
// format. The caller owns the returned msg memory.
//
// If there is an error, possible values are:
//   * io.EOF, when no messages remain
//   * io.ErrUnexpectedEOF
//   * of type transport.ConnectionError
//   * an error from the status package
// No other error values or types must be returned, which also means
// that the underlying io.Reader must not return an incompatible
// error.
func (p *grpcRequestParser) Parse() (pf GRPCPayloadCompressionFormat, msg []byte, err error) {
	if _, err := p.r.Read(p.header[:]); err != nil {
		return 0, nil, err
	}
	// first byte is for compressed or not
	pf = GRPCPayloadCompressionFormat(p.header[0])
	// next 4 bytes is for length of message
	length := binary.BigEndian.Uint32(p.header[1:])
	if length == 0 {
		return pf, nil, nil
	}
	if int64(length) > int64(maxInt) {
		return 0, nil, status.Errorf(codes.ResourceExhausted, "grpc: received message larger than max length allowed on current machine (%d vs. %d)", length, maxInt)
	}

	msg = make([]byte, int(length))
	if _, err := p.r.Read(msg); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return 0, nil, err
	}
	return pf, msg, nil
}

func fieldFromProtoMessage(msg []byte, tagIndex int) (string, error) {
	desc, err := buildPayloadGenericProto()
	if err != nil {
		return "", err
	}

	// populate message
	dynamicMsgKey := dynamic.NewMessage(desc)
	if err := dynamicMsgKey.Unmarshal(msg); err != nil {
		return "", err
	}
	val, err := dynamicMsgKey.TryGetFieldByNumber(tagIndex)
	if err != nil {
		return "", err
	}
	return val.(string), nil
}

// should only be built once
var genericProtoCache *desc.MessageDescriptor

func buildPayloadGenericProto() (*desc.MessageDescriptor, error) {
	if genericProtoCache != nil {
		return genericProtoCache, nil
	}

	builderMsg := builder.NewMessage("message")
	for i := 1; i < 100; i++ {
		builderMsg.AddField(builder.NewField(fmt.Sprintf("field_%d", i),
			builder.FieldTypeScalar(descriptor.FieldDescriptorProto_TYPE_STRING)).SetNumber(int32(i)))
	}
	desc, err := builderMsg.Build()
	if err != nil {
		return nil, err
	}
	genericProtoCache = desc
	return genericProtoCache, nil
}
