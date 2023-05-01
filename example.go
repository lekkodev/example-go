package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/lekkodev/go-sdk/client"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 3333, "port to serve http")
	flag.Parse()

	ctx := context.Background()
	lekko, closer := startLekko(ctx)
	defer closer(ctx)

	http.HandleFunc("/hello", serveHello(ctx, lekko))
	log.Printf("Serving http on port 127.0.0.1:%d!\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Error serving http: %v\n", err)
	}
}

func startLekko(ctx context.Context) (client.Client, client.CloseFunc) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	sidecarProvider, err := client.ConnectSidecarProvider(ctx, "https://localhost:50051", "test", &client.RepositoryKey{
		OwnerName: "lekkodev",
		RepoName:  "example",
	})
	if err != nil {
		log.Fatalf("Failed to start sidecar provider: %v\n", err)
	}
	return client.NewClient("default", sidecarProvider)
}

func serveHello(ctx context.Context, lekko client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		ctx = client.Add(ctx, "context-key", r.URL.Query().Get("context-key"))
		log.Printf("Got /hello request\n")
		suffix, err := lekko.GetString(ctx, "hello")
		if err != nil {
			log.Printf("Failed to read from lekko sidecar: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Connecting to lekko sidecar failed\n")
			return
		}
		io.WriteString(w, fmt.Sprintf("Hello, %s!\n", suffix))
	}
}
