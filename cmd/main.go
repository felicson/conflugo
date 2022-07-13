package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/felicson/conflugo"
	"github.com/felicson/conflugo/internal/confluence"
	"github.com/felicson/conflugo/internal/confluence/client"
)

var (
	ConfluenceLogin    string
	ConfluencePassword string
	ConfluenceSpace    string
	ConfluenceURL      string
)

func main() {
	ancestorFile := "confluence.ancestor"
	if _, err := os.Stat(ancestorFile); errors.Is(err, os.ErrNotExist) {
		return
	}
	ancestor, err := os.ReadFile(ancestorFile)
	if err != nil {
		log.Fatal(err)
	}

	c := confluence.Storage{
		SpaceKey: ConfluenceSpace,
		URL:      ConfluenceURL,
		Client:   client.NewBasicClient(ConfluenceLogin, ConfluencePassword),
	}

	if err := conflugo.Sync(context.TODO(), &c, string(ancestor)); err != nil {
		log.Fatal(err)
	}
}
