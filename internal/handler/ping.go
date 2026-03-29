package handler

import (
	"net/http"
)

func (h *URLHandler) pingHandler(w http.ResponseWriter, r *http.Request) {
	err := h.service.PingRepository()

	if err != nil {
		h.logger.Zap.Errorw("Ping failed", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pong"))
}
