package domain

import (
	"time"
)

type Ad struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	ImageUrl     string    `json:"image_url"`
	Placement    Placement `json:"placement"`
	Status       Status    `json:"status"`
	CreatedAt    time.Time `json:"created_at", default:"now"`
	DeactivateAt time.Time `json:"deactivate_at"`
	TTLMinutes   int       `json:"ttl_minutes"`
}

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

type Placement string

const (
	HomeScreen  Placement = "home_screen"
	RideSummary Placement = "ride_summary"
	MapView     Placement = "map_view"
)
