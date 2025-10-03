package http_server

import (
	"ads_backend/internal/ads_service"
	"ads_backend/internal/domain"
	"encoding/json"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type AdsHandler struct {
	log     *zap.Logger
	service ads_service.Service
}

func NewAdsHandler(log *zap.Logger, service ads_service.Service) *AdsHandler {
	return &AdsHandler{log: log, service: service}
}

func (h *AdsHandler) Pattern() string {
	return "/"
}

func (h *AdsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case r.Method == http.MethodPost && path == "/adposts":
		var req createAdRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := req.Validate()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ad := domain.Ad{
			Title:      req.Title,
			ImageUrl:   req.ImageURL,
			Placement:  req.Placement,
			TTLMinutes: *req.TTLMinutes,
		}

		created, err := h.service.CreateAd(ad)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(created)

	case r.Method == http.MethodGet && strings.HasPrefix(path, "/adposts/"):
		const prefix = "/adposts/"
		id := strings.TrimPrefix(path, prefix)
		if id == "" || strings.Contains(id, "/") {
			http.NotFound(w, r)
			return
		}

		ad, err := h.service.GetAd(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ad)

	case r.Method == http.MethodPost && strings.HasPrefix(path, "/adposts/") && strings.HasSuffix(path, "/deactivate"):
		const prefix = "/adposts/"
		const suffix = "/deactivate"
		pathWithoutPrefix := strings.TrimPrefix(path, prefix)
		id := strings.TrimSuffix(pathWithoutPrefix, suffix)

		if id == "" || strings.Contains(id, "/") {
			http.NotFound(w, r)
			return
		}

		ad, err := h.service.DeactivateAd(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ad)

	case r.Method == http.MethodGet && path == "/adspots":
		placementStr := r.URL.Query().Get("placement")
		statusStr := r.URL.Query().Get("status")

		if statusStr != "" && statusStr != "active" {
			http.Error(w, "status parameter must be 'active'", http.StatusBadRequest)
			return
		}

		if placementStr == "" {
			http.Error(w, "placement parameter is required", http.StatusBadRequest)
			return
		}

		placement := domain.Placement(placementStr)
		if placement != domain.HomeScreen && placement != domain.RideSummary && placement != domain.MapView {
			http.Error(w, "invalid placement value", http.StatusBadRequest)
			return
		}

		ads, err := h.service.ListEligibleActiveAdsByPlacement(placement)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ads)

	default:
		http.NotFound(w, r)
	}
}

func NewServeMux(routes []Route) *http.ServeMux {
	mux := http.NewServeMux()
	for _, route := range routes {
		mux.Handle(route.Pattern(), route)
	}
	return mux
}
