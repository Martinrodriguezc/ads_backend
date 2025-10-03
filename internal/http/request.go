package http_server

import (
	"ads_backend/internal/domain"
	"fmt"
)

type createAdRequest struct {
	Title      string           `json:"title"`
	ImageURL   string           `json:"imageUrl"`
	Placement  domain.Placement `json:"placement"`
	TTLMinutes *int             `json:"ttlMinutes,omitempty"`
}

func (r *createAdRequest) Validate() error {
	if r.Title == "" {
		return fmt.Errorf("title is required")
	}

	if r.ImageURL == "" {
		return fmt.Errorf("image_url is required")
	}

	if r.TTLMinutes != nil && *r.TTLMinutes < 0 {
		return fmt.Errorf("ttl_minutes must be greater than or equal to 0")
	}

	return nil
}
