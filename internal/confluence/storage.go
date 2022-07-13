package confluence

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
	"github.com/felicson/conflugo/internal/confluence/model"
)

const (
	contentTypeHeader = "Content-Type"
)

type Doer interface {
	Put(ctx context.Context, url string, headers http.Header, body io.Reader) ([]byte, error)
	Post(ctx context.Context, url string, headers http.Header, body io.Reader) ([]byte, error)
	Get(ctx context.Context, url string, headers http.Header) ([]byte, error)
}

// Storage present confluence layer.
type Storage struct {
	SpaceKey string
	URL      string
	Client   Doer
}

// UpdateDocument change document in storage.
func (s Storage) UpdateDocument(ctx context.Context, doc *model.Document, confl *internal.ConfluenceDocument, ancestorID string) error {
	pl := model.NewPayload(doc.Title, confl.Body, doc.Space.Key, ancestorID, doc.Version.Number+1)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&pl); err != nil {
		return fmt.Errorf("on marshall payload: %v", err)
	}

	if _, err := s.Client.Put(ctx, fmt.Sprintf("%s/rest/api/content/%s", s.URL, doc.ID), nil, &buf); err != nil {
		return fmt.Errorf("on do update request: %v", err)
	}
	if err := s.CreateAttachmentsByPath(ctx, doc.ID, confl.Prefix, confl.Attachments); err != nil {
		return fmt.Errorf("on create attachments: %v", err)
	}
	return nil
}

// CreateDocument make a new doc in storage.
func (s Storage) CreateDocument(ctx context.Context, title string, confl *internal.ConfluenceDocument, ancestorID string) (*model.Document, error) {
	pl := model.NewPayload(title, confl.Body, s.SpaceKey, ancestorID, 0)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&pl); err != nil {
		return nil, fmt.Errorf("on marshall payload: %v", err)
	}

	body, err := s.Client.Post(ctx, fmt.Sprintf("%s/rest/api/content", s.URL), nil, &buf)
	if err != nil {
		return nil, fmt.Errorf("on do request: %v", err)
	}
	var doc model.Document
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("on unmarshal document: %v", err)
	}

	if err := s.CreateAttachmentsByPath(ctx, doc.ID, confl.Prefix, confl.Attachments); err != nil {
		return nil, fmt.Errorf("on create attachments: %v", err)
	}
	return &doc, nil
}

// FindDocumentsByParent find docs by parentID and filter func.
func (s Storage) FindDocumentsByParent(ctx context.Context, parentID string, filter func(*model.Document) bool) ([]model.Document, error) {
	u := make(url.Values)
	u.Add("cql", fmt.Sprintf("space='%s' AND parent=%s", s.SpaceKey, parentID))
	u.Add("limit", "1000")
	u.Add("expand", "version,space")
	u.Add("start", "0")

	body, err := s.Client.Get(ctx, fmt.Sprintf("%s/rest/api/content/search?%s", s.URL, u.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("on do request: %v", err)
	}

	var cql model.Search
	if err := json.Unmarshal(body, &cql); err != nil {
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

// CreateAttachmentsByPath adds attachments to doc from paths.
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
	if _, err := s.Client.Post(ctx,
		fmt.Sprintf("%s/rest/api/content/%s/child/attachment?allowDuplicated=true", s.URL, parentID),
		headers,
		&b); err != nil {
		return fmt.Errorf("on do request: %v", err)
	}
	return nil
}

// DeleteDocumentByID not implemented.
func (s Storage) DeleteDocumentByID(_ context.Context) error {
	panic("not implemented yet")
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
