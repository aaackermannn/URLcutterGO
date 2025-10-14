package handler

import (
	"encoding/json"
	"net/http"
	"urlcutter/internal/models"
	"urlcutter/internal/service"

	"github.com/gorilla/mux"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{service: service}
}

// CreateShortURL создает короткую ссылку

func (h *Handler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var req models.CreateURLRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.service.CreateShortURL(req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

//Redirect перенапраавляет на оригинальный URL

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	short := vars["short"]

	original, err := h.service.Redirect(short)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, original, http.StatusFound)
}

// GetURLInfo возвращает информацию о короткой ссылке

func (h *Handler) GetURLInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	short := vars["short"]

	original, err := h.service.GetOriginalURL(short)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	resp := map[string]string{
		"original_url": original,
		"short_url":    short,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
