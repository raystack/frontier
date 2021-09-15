package cmd

import (
	"fmt"
	"github.com/odpf/shield/pipeline"
	"github.com/odpf/shield/proxy"
	"github.com/odpf/shield/store/local"
	"github.com/odpf/shield/structs"
	"github.com/spf13/afero"
	cli "github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	proxyDefaultAddr = 8080
	proxyTermChan    = make(chan os.Signal, 1)
)

// h2c reverse proxy
func proxyCommand() *cli.Command {
	var rulesDir string
	c := &cli.Command{
		Use:     "proxy",
		Short:   "Start proxy over h2c",
		Example: "shield serve proxy",
		RunE: func(c *cli.Command, args []string) error {

			ruleRepo := local.NewRuleRepository(afero.NewBasePathFs(afero.NewOsFs(), rulesDir))
			authorizers := []structs.Authorizer{
				// uncomment them to test
				//pipeline.BasicHeaderAuthorizer{},
				//pipeline.BasicGRPCPayloadAuthorizer{},
			}
			proxyHandler := proxy.NewHandler(pipeline.NewRegexMatcher(ruleRepo), authorizers)
			proxy := proxy.NewH2c(proxy.NewH2cRoundTripper(), proxyHandler)
			go func() {
				proxyURL := fmt.Sprintf(":%d", proxyDefaultAddr)
				fmt.Println("starting proxy at", proxyURL)
				mux := http.NewServeMux()
				mux.HandleFunc("/", proxy.Handle)

				//create a tcp listener
				proxyListener, err := net.Listen("tcp", proxyURL)
				if err != nil {
					log.Fatalf("failed to listen: %v", err)
				}

				proxySrv := http.Server{
					Addr:    proxyURL,
					Handler: h2c.NewHandler(mux, &http2.Server{}),
				}
				log.Fatal(proxySrv.Serve(proxyListener))
			}()

			// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
			signal.Notify(proxyTermChan, os.Interrupt)
			signal.Notify(proxyTermChan, os.Kill)
			signal.Notify(proxyTermChan, syscall.SIGTERM)

			// Block until we receive our signal.
			<-proxyTermChan

			return nil
		},
	}
	c.PersistentFlags().IntVar(&proxyDefaultAddr, "port", proxyDefaultAddr, "default port to start listening")
	c.PersistentFlags().StringVar(&rulesDir, "rule-dir", rulesDir, "full path to rules directory")
	c.MarkPersistentFlagRequired("rule-dir")
	return c
}
