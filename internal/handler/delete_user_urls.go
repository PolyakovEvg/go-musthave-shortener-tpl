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

	for _, id := range ids {
		h.deleter.Enqueue(service.DeleteTask{
			UserID: userID,
			IDs:    []string{id},
		})
	}

	w.WriteHeader(http.StatusAccepted)
}
