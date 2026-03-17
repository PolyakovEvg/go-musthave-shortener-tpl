package handler

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository/db"
	"net/http"
)

func (h *URLHandler) ping(w http.ResponseWriter, r *http.Request) {
	err := db.CheckConnection(h.config.DBDSN)

	if err != nil {
		http.Error(w, "cannot connect to the data base.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pong"))
}
