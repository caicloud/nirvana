package model

type Film struct {
	BaseModel
	Starships    []Identity `json:"starships"`
	Vehicles     []Identity `json:"vehicles"`
	Planets      []Identity `json:"planets"`
	Producer     string     `json:"producer"`
	Title        string     `json:"title"`
	Episode      Identity   `json:"episode"`
	Director     string     `json:"director"`
	OpeningCrawl string     `json:"opening_crawl"`
	Characters   []Identity `json:"characters"`
	Species      []Identity `json:"species"`
}
