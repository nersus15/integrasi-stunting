package entity

import "time"

type Satusehat struct {
	ID        string     `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type SatusehatInput struct {
	Name string `json:"name"`
}
