package model

import (
	"encoding/json"
	"strconv"

	"github.com/felicson/conflugo/internal/buf"
)

type Payload struct {
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Ancestors []Ancestor `json:"ancestors"`
	Space     Space      `json:"space"`
	Body      Body       `json:"body"`
	Version   Version    `json:"version"`
}
type Ancestor struct {
	ID int `json:"id"`
}
type Space struct {
	Key string `json:"key"`
}
type Body struct {
	Storage Storage `json:"storage"`
}
type Storage struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}
type Version struct {
	Number int `json:"number"`
}

func NewPayload(title string, value string, spaceKey, ancestorID string, version int) Payload {
	aID, _ := strconv.Atoi(ancestorID)
	return Payload{
		Type:      "page",
		Title:     title,
		Space:     Space{Key: spaceKey},
		Body:      Body{Storage: Storage{Value: value, Representation: "wiki"}},
		Version:   Version{Number: version},
		Ancestors: []Ancestor{{ID: aID}},
	}
}

func (pl *Payload) MarshalJSON() ([]byte, error) {
	type aliasPayload Payload
	pl2 := (*aliasPayload)(pl)
	var b buf.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(&pl2); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
