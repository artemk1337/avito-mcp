package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/artemk1337/avito-mcp/internal/avito"
	"github.com/artemk1337/avito-mcp/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	client, err := avito.NewClient(avito.BaseURL, nil, avito.Credentials{
		AccessToken:  os.Getenv("AVITO_ACCESS_TOKEN"),
		ClientID:     os.Getenv("AVITO_CLIENT_ID"),
		ClientSecret: os.Getenv("AVITO_CLIENT_SECRET"),
	})
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.New(client).Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
