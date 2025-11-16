package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/pigeio/todo-api/internal/middleware"
	"github.com/pigeio/todo-api/internal/models"
	"github.com/pigeio/todo-api/internal/repository" // Import for interface
	"github.com/pigeio/todo-api/internal/utils"
)

type TodoHandler struct {
	// Use your interface name
	todoRepo  repository.Todo_Repository
	validator *validator.Validate
}

// Use your interface name
func NewTodoHandler(todoRepo repository.Todo_Repository) *TodoHandler {
	return &TodoHandler{
		todoRepo:  todoRepo,
		validator: validator.New(),
	}
}

func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CreateTodoRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if err := h.validator.Struct(req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Title is required")
		return
	}

	// Create todo
	todo := &models.Todo{
		UserID:      claims.UserID,
		Title:       req.Title,
		Description: req.Description,
	}

	if err := h.todoRepo.Create(r.Context(), todo); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create todo")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, todo)
}

// --- THIS IS THE FIXED FUNCTION ---
func (h *TodoHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Read the new filter and sort parameters
	status := r.URL.Query().Get("status")
	sortBy := r.URL.Query().Get("sort_by")

	// --- THIS LINE WAS MISSING ---
	// It declares todos, total, and err, and uses claims, status, and sortBy
	todos, total, err := h.todoRepo.GetByUserID(r.Context(), claims.UserID, page, limit, status, sortBy)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch todos")
		return
	}
	// --- END OF FIX ---

	// Prepare response
	response := models.TodoListResponse{
		Data:  todos,
		Page:  page,
		Limit: limit,
		Total: total,
	}

	utils.RespondJSON(w, http.StatusOK, response)
}

func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get todo ID from URL
	vars := mux.Vars(r)
	todoID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}

	// Get existing todo
	todo, err := h.todoRepo.GetByID(r.Context(), todoID)
	if err != nil {
		utils.RespondError(w, http.StatusNotFound, "Todo not found")
		return
	}

	// Check authorization
	if todo.UserID != claims.UserID {
		utils.RespondError(w, http.StatusForbidden, "Forbidden")
		return
	}

	var req models.UpdateTodoRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update fields if provided
	if req.Title != "" {
		todo.Title = req.Title
	}
	if req.Description != "" {
		todo.Description = req.Description
	}
	if req.Completed != nil {
		todo.Completed = *req.Completed
	}

	// Update in database
	if err := h.todoRepo.Update(r.Context(), todo); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to update todo")
		return
	}

	utils.RespondJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get todo ID from URL
	vars := mux.Vars(r)
	todoID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}

	// Delete todo
	if err := h.todoRepo.Delete(r.Context(), todoID, claims.UserID); err != nil {
		utils.RespondError(w, http.StatusNotFound, "Todo not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
