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

type AIHandler struct {
	aiService   *service.AIService
	taskService *service.TaskService
}

func NewAIHandler(aiService *service.AIService, taskService *service.TaskService) *AIHandler {
	return &AIHandler{
		aiService:   aiService,
		taskService: taskService,
	}
}

// GenerateTasks creates tasks from natural language input
// POST /api/ai/generate
func (h *AIHandler) GenerateTasks(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req model.GenerateTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Generate tasks using AI
	generatedTasks, err := h.aiService.GenerateTasks(req.Input)
	if err != nil {
		if errors.Is(err, service.ErrAIUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "AI service is currently unavailable",
				"message": "Please try again later or create tasks manually",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate tasks",
			"message": err.Error(),
		})
		return
	}

	// Save generated tasks to database
	createdTasks, err := h.taskService.BulkCreateFromAI(userID, generatedTasks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save generated tasks",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.GenerateTasksResponse{
		Tasks:   createdTasks,
		Message: "Tasks generated successfully",
	})
}

// BreakdownTask breaks a task into subtasks
// POST /api/ai/breakdown/:id
func (h *AIHandler) BreakdownTask(c *gin.Context) {
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

	// Get the task to breakdown
	task, err := h.taskService.GetTask(taskID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
		return
	}

	// Generate subtasks using AI
	generatedSubtasks, err := h.aiService.BreakdownTask(task.Title, task.Description)
	if err != nil {
		if errors.Is(err, service.ErrAIUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "AI service is currently unavailable",
				"message": "Please try again later or create subtasks manually",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to breakdown task",
			"message": err.Error(),
		})
		return
	}

	// Save subtasks to database
	var createdSubtasks []model.Task
	for _, gen := range generatedSubtasks {
		// Use the parent task's category if not specified
		if gen.Category == "" || gen.Category == "Same as parent" {
			gen.Category = task.Category
		}

		subtask, err := h.taskService.CreateTaskFromAI(userID, gen, &taskID)
		if err != nil {
			continue // Skip failed subtasks
		}
		createdSubtasks = append(createdSubtasks, *subtask)
	}

	c.JSON(http.StatusCreated, model.BreakdownResponse{
		Subtasks: createdSubtasks,
		Message:  "Task broken down successfully",
	})
}

// SuggestPriority suggests a priority for a task
// POST /api/ai/suggest-priority
func (h *AIHandler) SuggestPriority(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req model.SuggestPriorityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	suggestion, err := h.aiService.SuggestPriority(req.Title, req.Description)
	if err != nil {
		if errors.Is(err, service.ErrAIUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "AI service is currently unavailable",
				"message": "Please try again later",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to suggest priority",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, suggestion)
}

// EstimateTime estimates the time for a task
// POST /api/ai/estimate-time
func (h *AIHandler) EstimateTime(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req model.EstimateTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	estimate, err := h.aiService.EstimateTime(req.Title, req.Description)
	if err != nil {
		if errors.Is(err, service.ErrAIUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "AI service is currently unavailable",
				"message": "Please try again later",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to estimate time",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, estimate)
}
