package routes

import (
	"go-sis-be/internal/handlers"
	"go-sis-be/middleware"

	"github.com/gorilla/mux"
)

func InitRouter() *mux.Router {
	r := mux.NewRouter()

	// Middleware Global
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.LoggingMiddleware)

	// Subrouter Utama /api/v1
	apiV1 := r.PathPrefix("/api/v1").Subrouter()

	// ===================================
	// A. Public Endpoints (TIDAK Butuh Token)
	// ===================================
	apiV1.HandleFunc("/login", handlers.LoginHandler).Methods("POST", "OPTIONS")
	apiV1.HandleFunc("/refresh", handlers.RefreshTokenHandler).Methods("POST", "OPTIONS")

	// ===================================
	// B. Protected Endpoints (Butuh Token)
	// ===================================
	// Terapkan AuthMiddleware pada semua endpoint di subrouter ini
	protectedRouter := apiV1.PathPrefix("").Subrouter()
	protectedRouter.Use(middleware.AuthMiddleware)

	// 1. Auth Maintenance
	protectedRouter.HandleFunc("/logout", handlers.LogoutHandler).Methods("POST", "OPTIONS") // <-- Hanya definisikan sekali

	// 2. User Management (CRUD)
	protectedRouter.HandleFunc("/users", handlers.CreateUserHandler).Methods("POST")
	protectedRouter.HandleFunc("/users", handlers.GetAllUsersHandler).Methods("GET")

	// Detail, Edit, Delete (UID)
	protectedRouter.HandleFunc("/users/{uid}", handlers.HandleGetUserDetail).Methods("GET")
	protectedRouter.HandleFunc("/users/{uid}", handlers.HandleEditProfile).Methods("PUT")
	protectedRouter.HandleFunc("/users/{uid}", handlers.HandleDeleteProfile).Methods("DELETE")

	// 3. Registrasi Spesifik (Role-specific creation)
	protectedRouter.HandleFunc("/register/student", handlers.HandleStudentRegistration).Methods("POST")
	protectedRouter.HandleFunc("/register/teacher", handlers.HandleTeacherRegistration).Methods("POST")
	protectedRouter.HandleFunc("/register/admin", handlers.HandleAdminRegistration).Methods("POST")
	protectedRouter.HandleFunc("/register/parent", handlers.HandleParentRegistration).Methods("POST")

	return r
}
