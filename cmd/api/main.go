package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/pigeio/todo-api/internal/database"
	"github.com/pigeio/todo-api/internal/handlers"
	"github.com/pigeio/todo-api/internal/middleware" // Import middleware
	"github.com/pigeio/todo-api/internal/repository"
	"github.com/pigeio/todo-api/internal/utils" // NEW IMPORT
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load JWT Secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set in .env file")
	}

	db, err := database.NewPostgresDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize REAL repositories
	userRepo := repository.NewUserRepository(db)
	todoRepo := repository.NewTodoRepository(db)

	// Initialize REAL Token Generator
	tokenGenerator, err := utils.NewJWTGenerator(jwtSecret)
	if err != nil {
		log.Fatal("Failed to create token generator:", err)
	}

	// Initialize Handlers, "plugging in" the real dependencies
	authHandler := handlers.NewAuthHandler(userRepo, tokenGenerator)

	// Note: You must also update NewTodoHandler to accept its interface
	todoHandler := handlers.NewTodoHandler(todoRepo)

	r := mux.NewRouter()
	r.HandleFunc("/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/login", authHandler.Login).Methods("POST")

	api := r.PathPrefix("/todos").Subrouter()

	// --- CHANGED ---
	// We now *call* AuthMiddleware, passing it the dependency it needs
	api.Use(middleware.AuthMiddleware(tokenGenerator))

	api.HandleFunc("", todoHandler.GetTodos).Methods("GET")
	api.HandleFunc("", todoHandler.CreateTodo).Methods("POST")
	api.HandleFunc("/{id}", todoHandler.UpdateTodo).Methods("PUT")
	api.HandleFunc("/{id}", todoHandler.DeleteTodo).Methods("DELETE")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Server shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
