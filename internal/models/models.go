package models

import (
	"time"
)

type URL struct {
	Id        string    `json:"id" db:"id"`
	Original  string    `json:"original_url" db:"original_url"`
	Short     string    `json:"short_url" db:"short_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Clicks    int       `json:"clicks" db:"clicks"`
}

type CreateURLRequest struct {
	URL string `json:"url" validate:"required, url"`
}

type CreateURLResponse struct {
	ShortURL string `json:"short_url"`
}
