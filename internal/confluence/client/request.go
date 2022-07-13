package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type Request struct {
	method  string
	url     string
	headers http.Header
	body    io.Reader
}

func (r *Request) String() string {
	var b bytes.Buffer
	_ = r.headers.WriteSubset(&b, excludedHeaders)
	return fmt.Sprintf("method: %s, url: %s, headers: %s", r.method, r.url, b.String())
}
