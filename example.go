package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
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
	fmt.Printf("Serving http on port 127.0.0.1:%d!\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Printf("Error serving http: %v\n", err)
		os.Exit(1)
	}
}

func startLekko(ctx context.Context) (client.Client, client.CloseFunc) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rk := &client.RepositoryKey{
		OwnerName: "lekkodev",
		RepoName:  "example",
	}
	var opts []client.ProviderOption
	provider, err := client.CachedGitFsProvider(ctx, rk, "../example/", opts...)
	if err != nil {
		fmt.Printf("Failed to start sidecar provider: %v\n", err)
		os.Exit(1)
	}
	return client.NewClient("default", provider)
}

func serveHello(ctx context.Context, lekko client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		urlVal := r.URL.Query().Get("context-key")
		if len(urlVal) > 0 {
			ctx = client.Add(ctx, "context-key", urlVal)
		}
		fmt.Printf("Got /hello request\n")
		suffix, err := lekko.GetString(ctx, "hello")
		if err != nil {
			fmt.Printf("Failed to read from lekko sidecar: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Connecting to lekko sidecar failed\n")
			return
		}
		io.WriteString(w, fmt.Sprintf("Hello, %s!\n", suffix))
	}
}
