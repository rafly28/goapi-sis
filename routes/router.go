package routes

import (
	"go-sis-be/handlers"
	"go-sis-be/middleware"

	"github.com/gorilla/mux"
)

func InitRouter() *mux.Router {
	r := mux.NewRouter()

	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.LoggingMiddleware)

	apiPublic := r.PathPrefix("/api/v1").Subrouter()
	apiPublic.HandleFunc("/login", handlers.LoginHandler).Methods("POST", "OPTIONS")
	apiPublic.HandleFunc("/refresh", handlers.RefreshTokenHandler).Methods("POST", "OPTIONS")

	apiProtected := r.PathPrefix("/api/v1").Subrouter()
	apiProtected.Use(middleware.AuthMiddleware)

	apiProtected.HandleFunc("/users", handlers.CreateUserHandler).Methods("POST")
	apiProtected.HandleFunc("/users", handlers.GetAllUsersHandler).Methods("GET")
	apiProtected.HandleFunc("/users/{uid}", handlers.GetUserHandler).Methods("GET")
	apiProtected.HandleFunc("/users/{uid}", handlers.DeleteUserHandler).Methods("DELETE")
	apiProtected.HandleFunc("/logout", handlers.LogoutHandler).Methods("POST")

	return r
}
