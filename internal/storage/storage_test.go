package storage

import (
	"context"
	"os"
	"testing"
)

func TestStorage_CreateAttachmentByPath(t *testing.T) {
	s := Storage{
		Login:    os.Getenv("CONFLUENCE_LOGIN"),
		Password: os.Getenv("CONFLUENCE_PASSWORD"),
		SpaceKey: os.Getenv("CONFLUENCE_SPACE"),
		URL:      os.Getenv("CONFLUENCE_URL"),
	}
	if err := s.CreateAttachmentsByPath(context.Background(), "111111", "",
		[]string{"image.png"}); err != nil {
		t.Fatal(err)
	}
}
