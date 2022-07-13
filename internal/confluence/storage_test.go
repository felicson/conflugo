package confluence

import (
	"context"
	"io"
	"net/http"
	"os"
	"testing"
)

type fake struct{}

func (f fake) Put(ctx context.Context, url string, headers http.Header, body io.Reader) ([]byte, error) {
	panic("implement me")
}

func (f fake) Post(ctx context.Context, url string, headers http.Header, body io.Reader) ([]byte, error) {
	panic("implement me")
}

func (f fake) Get(ctx context.Context, url string, headers http.Header) ([]byte, error) {
	panic("implement me")
}

func TestStorage_CreateAttachmentByPath(t *testing.T) {
	s := Storage{
		SpaceKey: os.Getenv("CONFLUENCE_SPACE"),
		URL:      os.Getenv("CONFLUENCE_URL"),
		Client:   fake{},
	}
	if err := s.CreateAttachmentsByPath(context.Background(), "111111", "",
		[]string{"image.png"}); err != nil {
		t.Fatal(err)
	}
}
