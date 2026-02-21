package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/task-manager/backend/internal/middleware"
	"github.com/task-manager/backend/internal/model"
	"github.com/task-manager/backend/internal/repository"
	"github.com/task-manager/backend/internal/service"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// ListTasks returns all tasks for the authenticated user
// GET /api/tasks
func (h *TaskHandler) ListTasks(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var filters model.TaskFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	tasks, err := h.taskService.ListTasks(userID, filters)
	if err != nil {
		if errors.Is(err, service.ErrInvalidStatus) || errors.Is(err, service.ErrInvalidPriority) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	c.JSON(http.StatusOK, model.TaskListResponse{Tasks: tasks})
}

// CreateTask creates a new task
// POST /api/tasks
func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req model.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	task, err := h.taskService.CreateTask(userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPriority) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// GetTask returns a single task by ID
// GET /api/tasks/:id
func (h *TaskHandler) GetTask(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	task, err := h.taskService.GetTask(taskID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateTask updates an existing task
// PUT /api/tasks/:id
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	var req model.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	task, err := h.taskService.UpdateTask(taskID, userID, &req)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidStatus) || errors.Is(err, service.ErrInvalidPriority) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask deletes a task
// DELETE /api/tasks/:id
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	err := h.taskService.DeleteTask(taskID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// GetCategories returns all categories for the user
// GET /api/tasks/categories
func (h *TaskHandler) GetCategories(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	categories, err := h.taskService.GetCategories(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}
