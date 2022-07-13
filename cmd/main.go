package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/felicson/conflugo"
	"github.com/felicson/conflugo/internal/storage"
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

	storage := storage.Storage{
		Login:    ConfluenceLogin,
		Password: ConfluencePassword,
		SpaceKey: ConfluenceSpace,
		URL:      ConfluenceURL,
	}

	if err := conflugo.Sync(context.TODO(), &storage, string(ancestor)); err != nil {
		log.Fatal(err)
	}
}
