package handler

import (
	authmw "PolyakovEvg/go-musthave-shortener-tpl/internal/middleware/auth"
	"encoding/json"
	"net/http"
)

func (h *URLHandler) getUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := authmw.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	urls, err := h.urlService.GetUserURLs(userID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type resp struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	result := make([]resp, 0, len(urls))

	for _, u := range urls {
		result = append(result, resp{
			ShortURL:    u.ShortURL,
			OriginalURL: u.OriginalURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
