package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/felicson/conflugo/internal"
	"github.com/felicson/conflugo/internal/storage/model"
)

const (
	contentTypeHeader = "Content-Type"
)

var excludedHeaders = map[string]bool{
	"Authorization": true,
}

// Storage present credentials for confluence
type Storage struct {
	Login    string
	Password string
	SpaceKey string
	URL      string
}

type request struct {
	method  string
	url     string
	headers http.Header
	body    io.Reader
}

type errResponse struct {
	Message string `json:"message"`
}

func (r *request) String() string {
	var b bytes.Buffer
	_ = r.headers.WriteSubset(&b, excludedHeaders)
	return fmt.Sprintf("method: %s, url: %s, headers: %s", r.method, r.url, b.String())
}

// UpdateDocument change document in storage
func (s Storage) UpdateDocument(ctx context.Context, doc *model.Document, confl *internal.ConfluenceDocument, ancestorID string) error {

	pl := model.NewPayload(doc.Title, string(confl.Body), doc.Space.Key, ancestorID, doc.Version.Number+1)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&pl); err != nil {
		return fmt.Errorf("on marshall payload: %v", err)
	}

	if _, err := s.doRequest(ctx, request{method: http.MethodPut, url: s.URL + "/rest/api/content/" + doc.ID, body: &buf}); err != nil {
		return fmt.Errorf("on do update request: %v", err)
	}
	if err := s.CreateAttachmentsByPath(ctx, doc.ID, confl.Prefix, confl.Attachments); err != nil {
		return fmt.Errorf("on create attachments: %v", err)
	}
	return nil
}

// CreateDocument make a new doc in storage
func (s Storage) CreateDocument(ctx context.Context, title string, confl *internal.ConfluenceDocument, ancestorID string) (*model.Document, error) {

	pl := model.NewPayload(title, confl.Body, s.SpaceKey, ancestorID, 0)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&pl); err != nil {
		return nil, fmt.Errorf("on marshall payload: %v", err)
	}

	b, err := s.doRequest(ctx, request{method: http.MethodPost, url: s.URL + "/rest/api/content", body: &buf})
	if err != nil {
		return nil, fmt.Errorf("on do request: %v", err)
	}
	var doc model.Document
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("on unmarshal document: %v", err)
	}

	if err := s.CreateAttachmentsByPath(ctx, doc.ID, confl.Prefix, confl.Attachments); err != nil {
		return nil, fmt.Errorf("on create attachments: %v", err)
	}
	return &doc, nil
}

// FindDocumentsByParent find docs by parentID and filter func
func (s Storage) FindDocumentsByParent(ctx context.Context, parentID string, filter func(*model.Document) bool) ([]model.Document, error) {

	u := url.Values{}
	u.Add("cql", fmt.Sprintf("space='%s' AND parent=%s", s.SpaceKey, parentID))
	u.Add("limit", "1000")
	u.Add("expand", "version,space")
	u.Add("start", "0")

	b, err := s.doRequest(ctx, request{method: http.MethodGet, url: s.URL + "/rest/api/content/search?" + u.Encode()})
	if err != nil {
		return nil, fmt.Errorf("on do request: %v", err)
	}

	var cql model.Search
	if err := json.Unmarshal(b, &cql); err != nil {
		return nil, fmt.Errorf("on unmarshal documents: %v", err)
	}
	var docs []model.Document
	for i, doc := range cql.Results {
		if filter(&cql.Results[i]) {
			docs = append(docs, doc)
		}
	}
	return docs, nil
}

// CreateAttachmentsByPath adds attachments to doc from paths
func (s Storage) CreateAttachmentsByPath(ctx context.Context, parentID string, prefix string, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	var (
		b bytes.Buffer
		// source document can contain multiply attachments with same name, it will cause confluence api error.
		unique = make(map[string]struct{})
	)
	mp := multipart.NewWriter(&b)
	for _, path := range paths {
		base := filepath.Base(path)
		if _, ok := unique[base]; ok {
			continue
		}
		unique[base] = struct{}{}
		w, err := mp.CreateFormFile("file", base)
		if err != nil {
			return fmt.Errorf("on create form file: %v", err)
		}
		filename := fmt.Sprintf("%s%s", prefix, path)
		if err := fileToWriter(filename, w); err != nil {
			return fmt.Errorf("on write file %s to writer: %v", filename, err)
		}
	}
	if err := mp.Close(); err != nil {
		return fmt.Errorf("on close multipart form: %v", err)
	}
	headers := make(http.Header)
	headers.Add("X-Atlassian-Token", "nocheck")
	headers.Add(contentTypeHeader, "multipart/form-data; boundary="+mp.Boundary())
	if _, err := s.doRequest(ctx,
		request{
			method:  http.MethodPost,
			url:     s.URL + "/rest/api/content/" + parentID + "/child/attachment?allowDuplicated=true",
			headers: headers,
			body:    &b}); err != nil {
		return fmt.Errorf("on do request: %v", err)
	}
	return nil
}

// DeleteDocumentByID not implemented
func (s Storage) DeleteDocumentByID(_ context.Context) error {
	panic("not implemented yet")
}

func (s *Storage) doRequest(ctx context.Context, request request) ([]byte, error) {

	req, err := http.NewRequestWithContext(ctx, request.method, request.url, request.body)
	if err != nil {
		return nil, fmt.Errorf("on new request %s: %v", request, err)
	}
	req.SetBasicAuth(s.Login, s.Password)
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

func fileToWriter(filename string, dst io.Writer) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("on open path %s: %v", filename, err)
	}
	defer f.Close()
	if _, err := io.Copy(dst, f); err != nil {
		return fmt.Errorf("on copy: %v", err)
	}
	return nil
}
