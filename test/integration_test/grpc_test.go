package integration_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/goto/shield/core/project"
	"github.com/goto/shield/core/rule"
	"github.com/goto/shield/internal/proxy"
	"github.com/goto/shield/internal/proxy/hook"
	"github.com/goto/shield/internal/store/blob"
	"github.com/goto/shield/test/integration_test/fixtures/helloworld"

	"github.com/goto/salt/log"
	"github.com/stretchr/testify/assert"

	"gocloud.dev/blob/fileblob"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"google.golang.org/grpc"
)

const (
	grpcBackendPort = 13877
	grpcProxyPort   = grpcBackendPort + 1
)

func TestGRPCProxyHelloWorld(t *testing.T) {
	baseCtx, baseCancel := context.WithCancel(context.Background())
	defer baseCancel()

	blobFS, err := fileblob.OpenBucket("./fixtures", &fileblob.Options{
		CreateDir: true,
		Metadata:  fileblob.MetadataDontWrite,
	})
	if err != nil {
		t.Fatal(err)
	}

	responseHooks := hookPipeline(log.NewNoop())
	h2cProxy := proxy.NewH2c(proxy.NewH2cRoundTripper(log.NewNoop(), responseHooks), proxy.NewDirector())
	ruleRepo := blob.NewRuleRepository(log.NewNoop(), blobFS)
	if err := ruleRepo.InitCache(baseCtx, time.Minute); err != nil {
		t.Fatal(err)
	}
	defer ruleRepo.Close()
	ruleService := rule.NewService(ruleRepo)
	projectService := project.Service{}
	pipeline := buildPipeline(log.NewNoop(), h2cProxy, ruleService, &projectService)

	proxyURL := fmt.Sprintf(":%d", grpcProxyPort)
	mux := http.NewServeMux()
	mux.Handle("/", pipeline)

	//create a tcp listener
	proxyListener, err := net.Listen("tcp", proxyURL)
	if err != nil {
		t.Fatal(err)
	}
	proxySrv := http.Server{
		Addr:    proxyURL,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}
	defer proxySrv.Close()
	go func() {
		if err := proxySrv.Serve(proxyListener); err != nil && err != http.ErrServerClosed {
			t.Error(err)
		}
	}()
	go func() {
		if err := startTestGRPCServer(grpcBackendPort, &greetServer{}); err != nil {
			t.Error(err)
		}
	}()

	// wait for proxy to start
	time.Sleep(time.Second * 1)

	t.Run("unary call with basic rpc credential and payload authorization", func(t *testing.T) {
		conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", grpcProxyPort),
			grpc.WithInsecure(),
			grpc.WithPerRPCCredentials(&BasicAuthentication{
				Token: "dXNlcjpwYXNzd29yZA==", // user:password
			}))
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		client := helloworld.NewGreeterClient(conn)
		resp, err := client.SayHello(context.Background(), &helloworld.HelloRequest{
			Name: "shield",
		})
		if err != nil {
			t.Error(err)
			assert.Nil(t, err)
		}
		assert.Equal(t, "Hello shield", resp.Message)
	})
	t.Run("stream call with basic rpc credential", func(t *testing.T) {
		conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", grpcProxyPort),
			grpc.WithInsecure(),
			grpc.WithPerRPCCredentials(&BasicAuthentication{
				Token: "dXNlcjpwYXNzd29yZA==", // user:password
			}))
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		client := helloworld.NewGreeterClient(conn)

		respStream, err := client.StreamExample(context.Background(), &helloworld.StreamExampleRequest{})
		if err != nil {
			t.Error(err)
			assert.Nil(t, err)
		}
		for {
			msg, err := respStream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Error(err)
			}
			assert.Contains(t, msg.Data, "foo-")
		}
	})
}

func BenchmarkGRPCProxyHelloWorld(b *testing.B) {
	baseCtx, baseCancel := context.WithCancel(context.Background())
	defer baseCancel()

	blobFS, err := fileblob.OpenBucket("./fixtures", &fileblob.Options{
		CreateDir: true,
		Metadata:  fileblob.MetadataDontWrite,
	})
	if err != nil {
		b.Fatal(err)
	}

	h2cProxy := proxy.NewH2c(proxy.NewH2cRoundTripper(log.NewNoop(), hook.New()), proxy.NewDirector())
	ruleRepo := blob.NewRuleRepository(log.NewNoop(), blobFS)
	if err := ruleRepo.InitCache(baseCtx, time.Minute); err != nil {
		b.Fatal(err)
	}
	defer ruleRepo.Close()
	ruleService := rule.NewService(ruleRepo)
	projectService := project.Service{}
	pipeline := buildPipeline(log.NewNoop(), h2cProxy, ruleService, &projectService)

	proxyURL := fmt.Sprintf(":%d", grpcProxyPort)
	mux := http.NewServeMux()
	mux.Handle("/", pipeline)

	//create a tcp listener
	proxyListener, err := net.Listen("tcp", proxyURL)
	if err != nil {
		b.Fatal(err)
	}
	proxySrv := http.Server{
		Addr:    proxyURL,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}
	defer proxySrv.Close()
	go func() {
		if err := proxySrv.Serve(proxyListener); err != nil && err != http.ErrServerClosed {
			b.Error(err)
		}
	}()
	go func() {
		if err := startTestGRPCServer(grpcBackendPort, &greetServer{}); err != nil {
			b.Error(err)
		}
	}()

	// wait for proxy to start
	time.Sleep(time.Second * 1)

	b.Run("unary call with basic rpc credential and payload authorization", func(b *testing.B) {
		conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", grpcProxyPort),
			grpc.WithInsecure(),
			grpc.WithPerRPCCredentials(&BasicAuthentication{
				Token: "dXNlcjpwYXNzd29yZA==", // user:password
			}))
		if err != nil {
			b.Fatal(err)
		}
		defer conn.Close()
		client := helloworld.NewGreeterClient(conn)
		for i := 0; i < b.N; i++ {
			resp, err := client.SayHello(context.Background(), &helloworld.HelloRequest{
				Name: "shield",
			})
			if err != nil {
				b.Error(err)
				assert.Nil(b, err)
			}
			assert.Equal(b, "Hello shield", resp.Message)
		}
	})
}

func startTestGRPCServer(port int, greetSrv helloworld.GreeterServer) error {
	s := grpc.NewServer()
	defer s.Stop()

	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		panic(err)
	}

	helloworld.RegisterGreeterServer(s, greetSrv)
	return s.Serve(lis)
}

type greetServer struct{}

func (s *greetServer) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{Message: "Hello " + in.Name}, nil
}

func (s *greetServer) StreamExample(in *helloworld.StreamExampleRequest, server helloworld.Greeter_StreamExampleServer) error {
	for i := 0; i < 10; i++ {
		if err := server.Send(&helloworld.StreamExampleReply{Data: fmt.Sprintf("foo-%d", i)}); err != nil {
			panic(err)
		}
	}
	return nil
}

type BasicAuthentication struct {
	Token string
}

func (a *BasicAuthentication) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": fmt.Sprintf("Basic %s", a.Token),
	}, nil
}

func (a *BasicAuthentication) RequireTransportSecurity() bool {
	return false
}
