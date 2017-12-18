package model

type Specie struct {
	BaseModel
	Classification  string     `json:"classification"`
	Designation     string     `json:"designation"`
	EyeColors       string     `json:"eye_colors"`
	People          []Identity `json:"people"`
	SkinColors      string     `json:"skin_colors"`
	Language        string     `json:"language"`
	Homeworld       *Identity  `json:"homeworld"`
	AverageLifespan string     `json:"average_lifespan"`
	AverageHeight   string     `json:"average_height"`
}
