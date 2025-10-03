package persistence

import (
	"fmt"
	"sync"
	"time"

	"ads_backend/internal/domain"
)

type AdRepository interface {
	CreateAd(ad domain.Ad) (domain.Ad, error)
	GetAd(id string) (domain.Ad, error)
	DeactivateAd(id string) (domain.Ad, error)
	ListEligibleActiveAdsByPlacement(placement domain.Placement) ([]domain.Ad, error)
}

type adRepository struct {
	ads map[string]domain.Ad
	mu  sync.RWMutex
}

func NewAdRepository() AdRepository {
	return &adRepository{
		ads: make(map[string]domain.Ad),
		mu:  sync.RWMutex{},
	}
}

func (r *adRepository) CreateAd(ad domain.Ad) (domain.Ad, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ads[ad.ID] = ad
	return ad, nil
}

func (r *adRepository) GetAd(id string) (domain.Ad, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ad := r.ads[id]
	if ad == (domain.Ad{}) {
		return domain.Ad{}, fmt.Errorf("ad not found")
	}
	return ad, nil
}

func (r *adRepository) DeactivateAd(id string) (domain.Ad, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ad := r.ads[id]
	if ad == (domain.Ad{}) {
		return domain.Ad{}, fmt.Errorf("ad not found")
	}
	ad.Status = domain.StatusInactive
	ad.DeactivateAt = time.Now()
	r.ads[id] = ad
	return ad, nil
}

func (r *adRepository) ListEligibleActiveAdsByPlacement(placement domain.Placement) ([]domain.Ad, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.Ad, 0)
	now := time.Now()

	for _, ad := range r.ads {
		if ad.Status != domain.StatusActive {
			continue
		}

		if ad.Placement != placement {
			continue
		}

		if ad.TTLMinutes > 0 {
			expirationTime := ad.CreatedAt.Add(time.Duration(ad.TTLMinutes) * time.Minute)
			if now.After(expirationTime) {
				continue
			}
		}

		result = append(result, ad)
	}

	return result, nil
}
