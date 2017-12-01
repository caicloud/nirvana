package model

type Transport struct {
	BaseModel
	Consumables          string `json:"consumables"`
	Name                 string `json:"name"`
	CargoCapacity        string `json:"cargo_capacity"`
	Passengers           string `json:"passengers"`
	MaxAtmospheringSpeed string `json:"max_atmosphering_speed"`
	Crew                 string `json:"crew"`
	Length               string `json:"length"`
	Model                string `json:"model"`
	CostInCredits        string `json:"cost_in_credits"`
	Manufacturer         string `json:"manufacturer"`
}
