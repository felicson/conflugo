package model

type Search struct {
	Results []Document `json:"results"`
}

type Document struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	Title   string `json:"title"`
	Version struct {
		Number int `json:"number"`
	} `json:"version"`
	Space struct {
		Key string `json:"key"`
	} `json:"space"`
	Links struct {
		WebUI  string `json:"webui"`
		TinyUI string `json:"tinyui"`
		Self   string `json:"self"`
	} `json:"_links"`
	Expandable struct {
		Version     string `json:"version"`
		Descendants string `json:"descendants"`
		Space       string `json:"space"`
	} `json:"_expandable"`
}
