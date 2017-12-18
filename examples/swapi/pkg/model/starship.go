package model

type Starship struct {
	Pilots            []Identity `json:"pilots"`
	MGLT              string     `json:"MGLT"`
	StartshipClass    string     `json:"startship_class"`
	HyperdriveRatting string     `json:"hyperdrive_ratting"`
}
