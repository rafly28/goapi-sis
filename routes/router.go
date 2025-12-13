package routes

import (
	"go-sis-be/internal/handlers"
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
	apiProtected.HandleFunc("/users/{uid}", handlers.HandleGetUserDetail).Methods("GET")
	apiProtected.HandleFunc("/users/{uid}", handlers.DeleteUserHandler).Methods("DELETE")
	apiProtected.HandleFunc("/logout", handlers.LogoutHandler).Methods("POST")
	apiProtected.HandleFunc("/register/student", handlers.HandleStudentRegistration).Methods("POST")
	apiProtected.HandleFunc("/register/teacher", handlers.HandleTeacherRegistration).Methods("POST")
	apiProtected.HandleFunc("/register/admin", handlers.HandleAdminRegistration).Methods("POST")
	apiProtected.HandleFunc("/register/parent", handlers.HandleParentRegistration).Methods("POST")

	return r
}
