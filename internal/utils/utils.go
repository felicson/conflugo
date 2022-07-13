package utils

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/speps/go-hashids/v2"
)

const (
	readme   = "README.md"
	lenOfExt = len(".md")
)

func MakeSuffix(ancestorID string) (string, error) {
	ancHex := hex.EncodeToString([]byte(ancestorID))

	hd := hashids.HashIDData{
		Alphabet:  "abcdefghijklmnopqrstuvwxyz1234567890",
		MinLength: 4,
		Salt:      ancestorID,
	}

	h, err := hashids.NewWithData(&hd)
	if err != nil {
		return "", fmt.Errorf("on new hash ids: %v", err)
	}
	suffix, err := h.EncodeHex(ancHex)
	if err != nil {
		return "", fmt.Errorf("on encode ancector hex: %v", err)
	}
	return "___" + suffix[:6], nil
}

func PackageName() string {
	return readme[:len(readme)-lenOfExt]
}

func TitleByPath(path string) string {
	name := filepath.Base(path)
	lenName := len(name)
	if lenName == 0 || lenName < lenName-lenOfExt {
		return "empty title"
	}
	name = name[:len(name)-lenOfExt]
	return strings.Join(strings.FieldsFunc(name, clean), " ")
}

func clean(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsNumber(r)
}
