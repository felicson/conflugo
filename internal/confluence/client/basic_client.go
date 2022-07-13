package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	contentTypeHeader = "Content-Type"
)

type errResponse struct {
	Message string `json:"message"`
}

var excludedHeaders = map[string]bool{
	"Authorization": true,
}

type Web struct {
	login    string
	password string
}

func (s Web) Get(ctx context.Context, url string, headers http.Header) ([]byte, error) {
	request := Request{
		url:     url,
		method:  http.MethodGet,
		headers: headers,
	}
	return s.do(ctx, request)
}

func (s Web) Put(ctx context.Context, url string, headers http.Header, body io.Reader) ([]byte, error) {
	request := Request{
		url:     url,
		method:  http.MethodPut,
		headers: headers,
		body:    body,
	}
	return s.do(ctx, request)
}

func (s Web) Post(ctx context.Context, url string, headers http.Header, body io.Reader) ([]byte, error) {
	request := Request{
		url:     url,
		method:  http.MethodPost,
		headers: headers,
		body:    body,
	}
	return s.do(ctx, request)
}

func (s Web) do(ctx context.Context, request Request) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, request.method, request.url, request.body)
	if err != nil {
		return nil, fmt.Errorf("on new request %s: %v", request, err)
	}
	req.SetBasicAuth(s.login, s.password)
	if request.headers != nil {
		for k, v := range request.headers {
			req.Header[k] = v
		}
	}
	// setting default content-type
	if _, ok := req.Header[contentTypeHeader]; !ok {
		req.Header.Add(contentTypeHeader, "application/json")
	}

	cl := http.DefaultClient
	resp, err := cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("on do %q request: %v", request, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var er errResponse
		if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
			return nil, fmt.Errorf("on decode error: %v", err)
		}
		return nil, fmt.Errorf("status not 200 for %s, %d - %s, reason: %s", request.String(), resp.StatusCode, resp.Status, er.Message)
	}
	return io.ReadAll(resp.Body)
}

func NewBasicClient(login, password string) Web {
	return Web{
		login:    login,
		password: password,
	}
}
