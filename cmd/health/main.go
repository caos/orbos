package main

import (
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/caos/orbos/internal/helpers"
)

func main() {

	listen := flag.String("listen", "", "Proxy health checks at this listen address")
	protocol := flag.String("protocol", "", "HTTP or HTTPS")
	ip := flag.String("ip", "", "Target IP")
	port := flag.Int("port", 0, "Target Port")
	path := flag.String("path", "", "Target Path")
	status := flag.Int("status", 0, "Expected response status")
	proxy := flag.Bool("proxy", false, "Test a proxy protocol using endpoint")

	flag.Parse()

	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	if *port > math.MaxUint16 {
		panic(fmt.Errorf("max port allowed: %d", math.MaxUint16))
	}

	listenVal := *listen
	if listenVal == "" {
		msg, err := helpers.Check(*protocol, *ip, uint16(*port), *path, *status, *proxy)
		if err != nil {
			panic(err)
		}
		fmt.Printf(msg)
		os.Exit(0)
	}

	pathStart := strings.Index(listenVal, "/")
	serve := listenVal[0:pathStart]
	servePath := listenVal[pathStart:]
	http.HandleFunc(servePath, func(writer http.ResponseWriter, request *http.Request) {
		msg, err := helpers.Check(*protocol, *ip, uint16(*port), *path, *status, *proxy)
		if err != nil {
			writer.WriteHeader(http.StatusServiceUnavailable)
		}
		writer.Write([]byte(msg))
		request.Body.Close()
	})
	fmt.Printf("Serving healthchecks at %s", listenVal)
	if err := http.ListenAndServe(serve, nil); err != nil {
		panic(err)
	}
}
