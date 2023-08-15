package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/lekkodev/go-sdk/client"
)

func main() {
	var port int
	var local bool
	var apiKey string
	flag.IntVar(&port, "port", 3333, "port to serve http")
	flag.BoolVar(&local, "local", false, "use local config repo")
	flag.StringVar(&apiKey, "apikey", "", "Lekko API key")
	flag.Parse()

	ctx := context.Background()
	lekko, closer := startLekko(ctx, local, apiKey)
	defer closer(ctx)

	http.HandleFunc("/hello", serveHello(ctx, lekko))
	fmt.Printf("Example app listening on port %d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Printf("Error serving http: %v\n", err)
		os.Exit(1)
	}
}

func startLekko(ctx context.Context, local bool, apiKey string) (client.Client, client.CloseFunc) {
	rk := &client.RepositoryKey{
		OwnerName: "lekkodev",
		RepoName:  "example",
	}
	var opts []client.ProviderOption
	if len(apiKey) > 0 {
		opts = append(opts, client.WithAPIKey(apiKey))	
	}
	var provider client.Provider
	var err error
	if local {
		provider, err = client.CachedGitFsProvider(ctx, rk, "../example", opts...)
	} else {
		provider, err = client.CachedAPIProvider(ctx, rk, opts...)
	}
	if err != nil {
		fmt.Printf("Failed to start Lekko provider: %v\n", err)
		os.Exit(1)
	}
	return client.NewClient("default", provider)
}

func serveHello(ctx context.Context, lekko client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		fmt.Printf("Got /hello request\n")
		urlVal := r.URL.Query().Get("context-key")
		if len(urlVal) > 0 {
			ctx = client.Add(ctx, "context-key", urlVal)
		}
		suffix, err := lekko.GetString(ctx, "hello")
		if err != nil {
			fmt.Printf("Failed to read from Lekko: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Failed to read from Lekko\n")
			return
		}
		io.WriteString(w, fmt.Sprintf("Hello, %s!\n", suffix))
	}
}
