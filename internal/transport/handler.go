package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/AnatoliySar/test/config"
	"github.com/AnatoliySar/test/internal/domain"
)

type Handler struct {
	broker domain.Broker
	cfg    *config.Config
}

func NewHandler(broker domain.Broker, cfg *config.Config) *Handler {
	return &Handler{
		broker: broker,
		cfg:    cfg,
	}
}

func (h *Handler) PutMessage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	queueName := r.URL.Path[len("/queue/"):]
	if queueName == "" {
		http.Error(w, "Queue name is required", http.StatusBadRequest)
		return
	}

	var request struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	if err := h.broker.Put(queueName, request.Message); err != nil {
		switch err {
		case domain.ErrQueueFull, domain.ErrQueueLimit:
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetMessage(w http.ResponseWriter, r *http.Request) {
	queueName := r.URL.Path[len("/queue/"):]
	if queueName == "" {
		http.Error(w, "Queue name is required", http.StatusBadRequest)
		return
	}

	// Парсим timeout
	timeout := h.cfg.DefaultTimeout
	if timeoutStr := r.URL.Query().Get("timeout"); timeoutStr != "" {
		var err error
		timeout, err = strconv.Atoi(timeoutStr)
		if err != nil || timeout < 0 {
			http.Error(w, "Invalid timeout value", http.StatusBadRequest)
			return
		}
	}

	msg, err := h.broker.Get(queueName, time.Duration(timeout)*time.Second)
	if err != nil {
		if err == domain.ErrMessageNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": msg.Data})
}
