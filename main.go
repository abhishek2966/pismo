package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/abhishek2966/pismo/pkg/handler"
	"github.com/abhishek2966/pismo/pkg/store/inmemory"
)

func main() {
	var (
		port        string
		clusterSize uint64
	)

	flag.StringVar(&port, "p", "80", "the service port")
	flag.Uint64Var(&clusterSize, "n", 10, "cluster size")
	flag.Parse()

	s := inmemory.InitDB(clusterSize)
	// initialize the handler with the logger that implements io.Writer interface
	h := handler.InitHandler(s, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /accounts", h.CreateAccount)
	mux.HandleFunc("GET /accounts/{accountId}", h.FetchAccount)
	mux.HandleFunc("POST /transactions", h.Transact)

	log.Printf("serving at port:%v", port)

	// start the server
	log.Fatal(http.ListenAndServe(fmt.Sprint(":", port), mux))
}
