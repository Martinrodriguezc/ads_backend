package http_server

import (
	"ads_backend/internal/domain"
	"ads_backend/mocks"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestListEligibleActiveAdsByPlacement_Success(t *testing.T) {
	mockAds := []domain.Ad{
		{
			ID:           "1",
			Title:        "Test Ad 1",
			ImageUrl:     "http://example.com/ad1.jpg",
			Placement:    domain.HomeScreen,
			Status:       domain.StatusActive,
			CreatedAt:    time.Now(),
			DeactivateAt: time.Now().Add(1 * time.Hour),
			TTLMinutes:   60,
		},
		{
			ID:           "2",
			Title:        "Test Ad 2",
			ImageUrl:     "http://example.com/ad2.jpg",
			Placement:    domain.HomeScreen,
			Status:       domain.StatusActive,
			CreatedAt:    time.Now(),
			DeactivateAt: time.Now().Add(2 * time.Hour),
			TTLMinutes:   120,
		},
	}

	mockService := mocks.NewService(t)
	mockService.On("ListEligibleActiveAdsByPlacement", domain.HomeScreen).
		Return(mockAds, nil)

	handler := NewAdsHandler(zap.NewNop(), mockService)

	req := httptest.NewRequest(http.MethodGet, "/adspots?placement=home_screen&status=active", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var ads []domain.Ad
	err := json.NewDecoder(w.Body).Decode(&ads)
	assert.NoError(t, err)
	assert.Len(t, ads, 2)
	assert.Equal(t, "Test Ad 1", ads[0].Title)
	assert.Equal(t, "Test Ad 2", ads[1].Title)

	mockService.AssertExpectations(t)
}

func TestListEligibleActiveAdsByPlacement_MissingPlacement(t *testing.T) {
	mockService := mocks.NewService(t)

	handler := NewAdsHandler(zap.NewNop(), mockService)

	req := httptest.NewRequest(http.MethodGet, "/adspots?status=active", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "placement parameter is required")

	mockService.AssertNotCalled(t, "ListEligibleActiveAdsByPlacement")
}

func TestListEligibleActiveAdsByPlacement_InvalidPlacement(t *testing.T) {
	mockService := mocks.NewService(t)

	handler := NewAdsHandler(zap.NewNop(), mockService)

	req := httptest.NewRequest(http.MethodGet, "/adspots?placement=invalid_placement&status=active", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid placement value")

	mockService.AssertNotCalled(t, "ListEligibleActiveAdsByPlacement")
}

func TestListEligibleActiveAdsByPlacement_InvalidStatus(t *testing.T) {
	mockService := mocks.NewService(t)

	handler := NewAdsHandler(zap.NewNop(), mockService)

	req := httptest.NewRequest(http.MethodGet, "/adspots?placement=home_screen&status=inactive", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "status parameter must be 'active'")

	mockService.AssertNotCalled(t, "ListEligibleActiveAdsByPlacement")
}

func TestListEligibleActiveAdsByPlacement_ServiceError(t *testing.T) {
	mockService := mocks.NewService(t)
	mockService.On("ListEligibleActiveAdsByPlacement", domain.RideSummary).
		Return([]domain.Ad{}, errors.New("database connection failed"))

	handler := NewAdsHandler(zap.NewNop(), mockService)

	req := httptest.NewRequest(http.MethodGet, "/adspots?placement=ride_summary&status=active", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "database connection failed")

	mockService.AssertExpectations(t)
}

func TestListEligibleActiveAdsByPlacement_EmptyResult(t *testing.T) {
	mockService := mocks.NewService(t)
	mockService.On("ListEligibleActiveAdsByPlacement", domain.MapView).
		Return([]domain.Ad{}, nil)

	handler := NewAdsHandler(zap.NewNop(), mockService)

	req := httptest.NewRequest(http.MethodGet, "/adspots?placement=map_view&status=active", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var ads []domain.Ad
	err := json.NewDecoder(w.Body).Decode(&ads)
	assert.NoError(t, err)
	assert.Len(t, ads, 0)

	mockService.AssertExpectations(t)
}

func TestListEligibleActiveAdsByPlacement_WithoutStatusParam(t *testing.T) {
	mockAds := []domain.Ad{
		{
			ID:        "1",
			Title:     "Test Ad",
			Placement: domain.HomeScreen,
			Status:    domain.StatusActive,
		},
	}

	mockService := mocks.NewService(t)
	mockService.On("ListEligibleActiveAdsByPlacement", domain.HomeScreen).
		Return(mockAds, nil)

	handler := NewAdsHandler(zap.NewNop(), mockService)

	req := httptest.NewRequest(http.MethodGet, "/adspots?placement=home_screen", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var ads []domain.Ad
	err := json.NewDecoder(w.Body).Decode(&ads)
	assert.NoError(t, err)
	assert.Len(t, ads, 1)

	mockService.AssertExpectations(t)
}

func TestCreateAd_Success(t *testing.T) {
	ttl := 60
	requestBody := map[string]interface{}{
		"title":      "New Ad",
		"imageUrl":   "http://example.com/ad.jpg",
		"placement":  "home_screen",
		"ttlMinutes": ttl,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	expectedAd := domain.Ad{
		ID:           "generated-id",
		Title:        "New Ad",
		ImageUrl:     "http://example.com/ad.jpg",
		Placement:    domain.HomeScreen,
		Status:       domain.StatusActive,
		CreatedAt:    time.Now(),
		DeactivateAt: time.Now().Add(time.Hour),
		TTLMinutes:   60,
	}

	mockService := mocks.NewService(t)
	mockService.On("CreateAd", mock.MatchedBy(func(ad domain.Ad) bool {
		return ad.Title == "New Ad" &&
			ad.ImageUrl == "http://example.com/ad.jpg" &&
			ad.Placement == domain.HomeScreen &&
			ad.TTLMinutes == 60
	})).Return(expectedAd, nil)

	handler := NewAdsHandler(zap.NewNop(), mockService)

	req := httptest.NewRequest(http.MethodPost, "/adposts", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var createdAd domain.Ad
	err := json.NewDecoder(w.Body).Decode(&createdAd)
	assert.NoError(t, err)
	assert.Equal(t, "generated-id", createdAd.ID)
	assert.Equal(t, "New Ad", createdAd.Title)

	mockService.AssertExpectations(t)
}

func TestGetAd_Success(t *testing.T) {
	expectedAd := domain.Ad{
		ID:           "123",
		Title:        "Existing Ad",
		ImageUrl:     "http://example.com/ad.jpg",
		Placement:    domain.RideSummary,
		Status:       domain.StatusActive,
		CreatedAt:    time.Now(),
		DeactivateAt: time.Now().Add(2 * time.Hour),
		TTLMinutes:   120,
	}

	mockService := mocks.NewService(t)
	mockService.On("GetAd", "123").Return(expectedAd, nil)

	handler := NewAdsHandler(zap.NewNop(), mockService)

	req := httptest.NewRequest(http.MethodGet, "/adposts/123", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var ad domain.Ad
	err := json.NewDecoder(w.Body).Decode(&ad)
	assert.NoError(t, err)
	assert.Equal(t, "123", ad.ID)
	assert.Equal(t, "Existing Ad", ad.Title)
	assert.Equal(t, domain.RideSummary, ad.Placement)

	mockService.AssertExpectations(t)
}

func TestDeactivateAd_Success(t *testing.T) {
	deactivatedAd := domain.Ad{
		ID:           "456",
		Title:        "Ad to Deactivate",
		ImageUrl:     "http://example.com/ad.jpg",
		Placement:    domain.MapView,
		Status:       domain.StatusInactive,
		CreatedAt:    time.Now().Add(-1 * time.Hour),
		DeactivateAt: time.Now(),
		TTLMinutes:   60,
	}

	mockService := mocks.NewService(t)
	mockService.On("DeactivateAd", "456").Return(deactivatedAd, nil)

	handler := NewAdsHandler(zap.NewNop(), mockService)

	req := httptest.NewRequest(http.MethodPost, "/adposts/456/deactivate", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var ad domain.Ad
	err := json.NewDecoder(w.Body).Decode(&ad)
	assert.NoError(t, err)
	assert.Equal(t, "456", ad.ID)
	assert.Equal(t, domain.StatusInactive, ad.Status)

	mockService.AssertExpectations(t)
}
