package handler

import (
	authmw "PolyakovEvg/go-musthave-shortener-tpl/internal/middleware/auth"
	service "PolyakovEvg/go-musthave-shortener-tpl/internal/service/deleter"
	"encoding/json"
	"net/http"
)

func (h *URLHandler) deleteUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := authmw.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var ids []string
	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	task := service.DeleteTask{
		UserID: userID,
		IDs:    ids,
	}

	if err := h.deleter.Enqueue(task); err != nil {
		h.logger.Zap.Warnw("failed to enqueue delete task",
			"userID", userID,
			"ids_count", len(ids),
			"error", err,
		)

		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
