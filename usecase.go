package conflugo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/felicson/conflugo/internal"
	"github.com/felicson/conflugo/internal/confluence"
	"github.com/felicson/conflugo/internal/confluence/model"
	"github.com/felicson/conflugo/internal/utils"
)

type doc struct {
	Path  string
	Title string
}

const (
	readme = "README.md"
)

// Sync entrypoint to use case.
func Sync(ctx context.Context, storage *confluence.Storage, ancestorID string) error {
	ancestorID = strings.Trim(ancestorID, " \n")

	if _, err := os.Stat(readme); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	suffix, err := utils.MakeSuffix(ancestorID)
	if err != nil {
		return err
	}

	packageName := utils.PackageName() + suffix

	findParent := func(d *model.Document) bool {
		return d.Title == packageName
	}
	findChild := func(d *model.Document) bool {
		return d.Title != packageName
	}
	var (
		document      *model.Document
		existingChild []model.Document
	)

	// first step finding a root node in defined ancestor tree
	docs, err := storage.FindDocumentsByParent(ctx, ancestorID, findParent)
	if err != nil {
		return fmt.Errorf("on find parent document: %v", err)
	}
	confDoc, err := internal.WikiByPath(readme)
	if err != nil {
		return fmt.Errorf("on make wiki page: %v", err)
	}
	if len(docs) == 0 {
		if _, err = storage.CreateDocument(ctx, packageName, confDoc, ancestorID); err != nil {
			return fmt.Errorf("on create %q as parent document: %v", packageName, err)
		}
	} else {
		// confluence document already exist do update
		document = &docs[0]
		if err = storage.UpdateDocument(ctx, document, confDoc, ancestorID); err != nil {
			return fmt.Errorf("on update %q as parent document: %v", document.Title, err)
		}
		existingChild, err = storage.FindDocumentsByParent(ctx, ancestorID, findChild)
		if err != nil {
			return err
		}
	}
	// step two handle child documents
	childReadme, err := handleDocDirectory(suffix)

	if err != nil {
		return fmt.Errorf("on handle doc dir: %v", err)
	}
	// update existing child documents
	for i, doc := range existingChild {
		readme, ok := childReadme[doc.Title]
		if !ok {
			// place for delete existing document logic in confluence
			continue
		}
		confDoc, err := internal.WikiByPath(readme.Path)
		if err != nil {
			return err
		}
		if err := storage.UpdateDocument(ctx, &existingChild[i], confDoc, ancestorID); err != nil {
			return fmt.Errorf("on update child doc: %s, %v", doc.Title, err)
		}
		delete(childReadme, doc.Title)
	}

	// creating a new child documents
	for _, readme := range childReadme {
		confDoc, err := internal.WikiByPath(readme.Path)
		if err != nil {
			return err
		}
		if _, err := storage.CreateDocument(ctx, readme.Title, confDoc, ancestorID); err != nil {
			return err
		}
	}

	return nil
}

func handleDocDirectory(suffix string) (map[string]doc, error) {
	docs, err := filepath.Glob("doc/*.md")
	if err != nil {
		return nil, err
	}
	var result = make(map[string]doc)
	for _, filePath := range docs {
		title := utils.TitleByPath(filePath) + suffix
		result[title] = doc{Path: filePath, Title: title}
	}
	return result, nil
}
