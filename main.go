package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

type Server struct {
	mux *chi.Mux
}

func NewServer(lc fx.Lifecycle, mux *chi.Mux, logger *zap.Logger) *Server {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				err := http.ListenAndServe(":8080", mux)
				if err != nil {
					logger.Error("Error starting HTTP server", zap.Error(err))
				}
				logger.Info("HTTP server started at :8080")
			}()
			return nil
		},
	})

	server := &Server{mux: mux}

	return server
}

const (
	_ = iota
	GET
	POST
	PUT
	DELETE
)

type Route interface {
	http.Handler

	Pattern() string
	Method() int
}

type HelloHandler struct {
	logger *zap.Logger
}

func NewHelloHandler(logger *zap.Logger) *HelloHandler {
	return &HelloHandler{logger: logger}
}

func (h *HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start: Hello, World!")
	w.Write([]byte("Hello, World!"))
	h.logger.Info("End: Hello, World!")
}

func (h *HelloHandler) Pattern() string {
	return "/hello"
}

func (h *HelloHandler) Method() int {
	return GET
}

type ByeHandler struct {
	logger *zap.Logger
}

func NewByeHandler(logger *zap.Logger) *ByeHandler {
	return &ByeHandler{logger: logger}
}

func (h *ByeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start: Bye, World!")
	w.Write([]byte("Bye, World!"))
	h.logger.Info("End: Bye, World!")
}

func (h *ByeHandler) Pattern() string {
	return "/bye"
}

func (h *ByeHandler) Method() int {
	return GET
}

func NewServeMux(routes []Route) *chi.Mux {
	mux := chi.NewMux()
	for _, route := range routes {
		switch route.Method() {
		case GET:
			mux.Get(route.Pattern(), route.ServeHTTP)
		case POST:
			mux.Post(route.Pattern(), route.ServeHTTP)
		case PUT:
			mux.Put(route.Pattern(), route.ServeHTTP)
		case DELETE:
			mux.Delete(route.Pattern(), route.ServeHTTP)
		}
	}
	return mux
}

func AsRoute(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(Route)),
		fx.ResultTags(`group:"routes"`),
	)
}

func main() {
	app := fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Provide(
			NewServer,
			fx.Annotate(
				NewServeMux,
				fx.ParamTags(`group:"routes"`),
			),
			AsRoute(NewHelloHandler),
			AsRoute(NewByeHandler),
			zap.NewExample,
		),
		fx.Invoke(func(*Server) {}),
	)

	app.Run()
}
