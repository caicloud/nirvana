package model

import "time"

type Identity int64

type BaseModel struct {
	Id      Identity  `json:"id"`
	Created time.Time `json:"created"`
	Edited  time.Time `json:"edited"`
}
