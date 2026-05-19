package handler

import (
	authmw "PolyakovEvg/go-musthave-shortener-tpl/internal/middleware/auth"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/model"
	"encoding/json"
	"io"
	"net/http"
)

func (h *URLHandler) shortenBatch(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Zap.Errorw("failed to read request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var batch []model.BatchRequest
	if err := json.Unmarshal(body, &batch); err != nil {
		h.logger.Zap.Errorw("invalid JSON format", "error", err)
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	if len(batch) == 0 {
		h.logger.Zap.Warnw("empty batch request")
		http.Error(w, "empty batch", http.StatusBadRequest)
		return
	}

	userID, ok := authmw.UserIDFromContext(r.Context())

	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	responses, err := h.urlService.SaveBatch(userID, batch)
	if err != nil {
		h.logger.Zap.Errorw("failed to save batch", "error", err, "batch_size", len(batch))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(responses); err != nil {
		h.logger.Zap.Errorw("failed to encode batch response", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
