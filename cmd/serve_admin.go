package cmd

import (
	"fmt"
	cli "github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	adminDefaultAddr = 8081
	termChan         = make(chan os.Signal, 1)
)

func adminCommand() *cli.Command {
	c := &cli.Command{
		Use:     "admin",
		Short:   "Start serving admin endpoint",
		Example: "shield serve admin",
		RunE: func(c *cli.Command, args []string) error {
			// TODO admin api

			//create a tcp listener for grpc
			lis, err := net.Listen("tcp", fmt.Sprintf(":%d", adminDefaultAddr))
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}

			// base router
			baseMux := http.NewServeMux()
			baseMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "pong")
				fmt.Println("pinged with headers", r.Header)
				if r.Body != nil {
					reqBody, _ := ioutil.ReadAll(r.Body)
					r.Body.Close()
					fmt.Println("pinged with body", string(reqBody))
				}
			})
			srv := &http.Server{
				Addr:         lis.Addr().String(),
				Handler:      baseMux,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  120 * time.Second,
			}
			go func() {
				fmt.Println("starting listening at", srv.Addr)
				if err := srv.Serve(lis); err != nil {
					if err != http.ErrServerClosed {
						panic(err)
					}
				}
			}()

			// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
			signal.Notify(termChan, os.Interrupt)
			signal.Notify(termChan, os.Kill)
			signal.Notify(termChan, syscall.SIGTERM)

			// Block until we receive our signal.
			<-termChan

			return nil
		},
	}
	c.PersistentFlags().IntVar(&adminDefaultAddr, "port", adminDefaultAddr, "default port to start listening")
	return c
}