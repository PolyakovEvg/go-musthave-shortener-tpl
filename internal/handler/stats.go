// Package handler предоставляет HTTP-хендлеры для сервиса сокращения URL.
package handler

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository"
	"encoding/json"
	"net"
	"net/http"
)

// StatsHandler обрабатывает запросы к эндпоинту статистики.
type StatsHandler struct {
	repo          repository.Repository
	trustedSubnet *net.IPNet // распарсенная доверенная подсеть
}

// NewStatsHandler создаёт новый экземпляр StatsHandler.
// Если trustedSubnet пустая строка или невалидный CIDR, trustedSubnet будет nil.
func NewStatsHandler(repo repository.Repository, trustedSubnet string) *StatsHandler {
	h := &StatsHandler{
		repo: repo,
	}

	if trustedSubnet != "" {
		_, ipNet, err := net.ParseCIDR(trustedSubnet)
		if err == nil {
			h.trustedSubnet = ipNet
		}
	}

	return h
}

// statsResponse представляет ответ эндпоинта /api/internal/stats.
// URLs — количество сокращённых URL в сервисе.
// Users — количество пользователей в сервисе.
type statsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// GetStats возвращает статистику сервиса.
// Доступ разрешён только для запросов из доверенной подсети.
// Если trusted_subnet не задан, доступ запрещён для всех.
func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if h.trustedSubnet == nil {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP == "" {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	ip := net.ParseIP(realIP)
	if ip == nil {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	if !h.trustedSubnet.Contains(ip) {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	stats, err := h.repo.GetStats()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := statsResponse{
		URLs:  stats.URLs,
		Users: stats.Users,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
