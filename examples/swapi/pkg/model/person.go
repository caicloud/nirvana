package model

type Gender string

type Person struct {
	BaseModel
	Name      string   `json:"name"`
	Gender    Gender   `json:"gender"`
	SkinColor string   `json:"skin_color"`
	HairColor string   `json:"hair_color"`
	Height    string   `json:"height"`
	EyeColor  string   `json:"eye_color"`
	Mass      string   `json:"mass"`
	Homeworld Identity `json:"homeworld"`
	BirthYear string   `json:"birth_year"`
}
