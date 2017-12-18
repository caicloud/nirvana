package model

type Planet struct {
	BaseModel
	Climate        string `json:"climate"`
	Name           string `json:"name"`
	Diameter       string `json:"diameter"`
	RotationPeriod string `json:"rotation_period"`
	Terrain        string `json:"terrain"`
	Gravity        string `json:"gravity"`
	OrbitalPeriod  string `json:"orbital_period"`
	Population     string `json:"population"`
}
