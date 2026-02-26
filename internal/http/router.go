package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/nate/go-boilerplate/internal/http/handler"
	httpmiddleware "github.com/nate/go-boilerplate/internal/http/middleware"
	"github.com/nate/go-boilerplate/internal/service"
	"github.com/nate/go-boilerplate/pkg/jwt"
)

type Router struct {
	*chi.Mux
}

func NewRouter(
	authService *service.AuthService,
	userService *service.UserService,
	noteService *service.NoteService,
	jwtManager *jwt.Manager,
	rateLimiter *httpmiddleware.RateLimiter,
) *Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httpmiddleware.Logger)
	r.Use(httpmiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(httpmiddleware.RateLimit(rateLimiter))

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	noteHandler := handler.NewNoteHandler(noteService)

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
		})

		r.Group(func(r chi.Router) {
			r.Use(httpmiddleware.Auth(jwtManager))
			r.Post("/logout-all", authHandler.LogoutAll)
		})

		r.Group(func(r chi.Router) {
			r.Use(httpmiddleware.Auth(jwtManager))
			r.Use(httpmiddleware.Authorize("admin"))

			r.Get("/users", userHandler.List)
			r.Delete("/users/{id}", userHandler.Delete)
		})

		r.Group(func(r chi.Router) {
			r.Use(httpmiddleware.Auth(jwtManager))

			r.Get("/notes", noteHandler.List)
			r.Post("/notes", noteHandler.Create)
			r.Get("/notes/{id}", noteHandler.Get)
			r.Put("/notes/{id}", noteHandler.Update)
			r.Delete("/notes/{id}", noteHandler.Delete)
			r.Post("/notes/{id}/restore", noteHandler.Restore)
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return &Router{r}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Mux.ServeHTTP(w, req)
}
