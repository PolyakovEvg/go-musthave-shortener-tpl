package handler

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/model"
	"encoding/json"
	"io"
	"net/http"
)

func (h *URLHandler) ShortenBatch(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var batch []model.BatchRequest
	if err := json.Unmarshal(body, &batch); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if len(batch) == 0 {
		http.Error(w, "empty batch", http.StatusBadRequest)
		return
	}

	responses, err := h.service.SaveBatch(batch)
	if err != nil {
		http.Error(w, "failed to save URLs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(responses); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
