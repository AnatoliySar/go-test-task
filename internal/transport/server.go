package transport

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/AnatoliySar/test/config"
	"github.com/AnatoliySar/test/internal/domain"
	broker_impl "github.com/AnatoliySar/test/internal/domain/impl"
)

var ErrTimeoutExceeded = errors.New("shutdown timeout exceeded")

type Server struct {
	broker domain.Broker
	server *http.Server
}

func NewServer(cfg *config.Config) *Server {
	broker := broker_impl.NewBroker(cfg.MaxQueues, cfg.MaxMessages)
	handler := NewHandler(broker, cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/queue/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			handler.PutMessage(w, r)
		case http.MethodGet:
			handler.GetMessage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return &Server{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: mux,
		},
		broker: broker,
	}
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.server.Shutdown(ctx)

	// Сюда можно записать логику для сохранения данных на диск перед выходом и т.д.

	return err
}
