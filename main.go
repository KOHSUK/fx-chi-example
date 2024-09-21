package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

// NewServer builds an HTTP server that will begin serving requests
// when the Fx application starts.
func NewServer(lc fx.Lifecycle) *chi.Mux {
	r := chi.NewRouter()

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				fmt.Println("Starting HTTP server at :8080")
				http.ListenAndServe(":8080", r)
			}()
			return nil
		},
	})

	return r
}

type HelloHandler struct{}

func NewHelloHandler() *HelloHandler {
	return &HelloHandler{}
}

func (h *HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func RegisterHelloHandler(mux *chi.Mux, handler *HelloHandler) {
	mux.Get("/hello", handler.ServeHTTP)
}

func main() {
	app := fx.New(
		fx.Provide(
			NewServer,
			NewHelloHandler,
		),
		fx.Invoke(RegisterHelloHandler),
	)

	app.Run()
}
