package ads_service

import (
	"ads_backend/internal/domain"
	"ads_backend/internal/persistence"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	CreateAd(ad domain.Ad) (domain.Ad, error)
	GetAd(id string) (domain.Ad, error)
	DeactivateAd(id string) (domain.Ad, error)
	ListEligibleActiveAdsByPlacement(placement domain.Placement) ([]domain.Ad, error)
}

type service struct {
	adRepository persistence.AdRepository
}

func NewService() Service {
	return &service{
		adRepository: persistence.NewAdRepository(),
	}
}

func (s *service) CreateAd(ad domain.Ad) (domain.Ad, error) {
	ad.ID = uuid.New().String()
	ad.Status = domain.StatusActive
	ad.CreatedAt = time.Now()
	if ad.TTLMinutes > 0 {
		ad.DeactivateAt = time.Now().Add(time.Duration(ad.TTLMinutes) * time.Minute)
	}
	ad, err := s.adRepository.CreateAd(ad)
	if err != nil {
		return domain.Ad{}, err
	}

	return ad, nil
}
func (s *service) GetAd(id string) (domain.Ad, error) {
	ad, err := s.adRepository.GetAd(id)
	if err != nil {
		return domain.Ad{}, err
	}

	return ad, nil
}
func (s *service) DeactivateAd(id string) (domain.Ad, error) {
	ad, err := s.adRepository.DeactivateAd(id)
	if err != nil {
		return domain.Ad{}, err
	}

	return ad, nil
}
func (s *service) ListEligibleActiveAdsByPlacement(placement domain.Placement) ([]domain.Ad, error) {
	ads, err := s.adRepository.ListEligibleActiveAdsByPlacement(placement)
	if err != nil {
		return []domain.Ad{}, err
	}
	return ads, nil
}
