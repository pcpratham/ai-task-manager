package service

import (
	"errors"
	"time"

	"github.com/task-manager/backend/internal/model"
	"github.com/task-manager/backend/internal/repository"
)

var (
	ErrInvalidStatus   = errors.New("invalid status: must be todo, in_progress, or done")
	ErrInvalidPriority = errors.New("invalid priority: must be low, medium, high, or urgent")
)

type TaskService struct {
	taskRepo *repository.TaskRepository
}

func NewTaskService(taskRepo *repository.TaskRepository) *TaskService {
	return &TaskService{taskRepo: taskRepo}
}

func (s *TaskService) CreateTask(userID string, req *model.CreateTaskRequest) (*model.Task, error) {
	// Validate priority if provided
	if req.Priority != "" && !isValidPriority(req.Priority) {
		return nil, ErrInvalidPriority
	}

	task := &model.Task{
		UserID:           userID,
		Title:            req.Title,
		Description:      req.Description,
		Status:           "todo",
		Priority:         req.Priority,
		Category:         req.Category,
		EstimatedMinutes: req.EstimatedMinutes,
		AIGenerated:      false,
	}

	if task.Priority == "" {
		task.Priority = "medium"
	}

	// Parse due date if provided
	if req.DueDate != nil && *req.DueDate != "" {
		dueDate, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return nil, errors.New("invalid due date format: use YYYY-MM-DD")
		}
		task.DueDate = &dueDate
	}

	// Set parent task ID if provided
	if req.ParentTaskID != nil && *req.ParentTaskID != "" {
		task.ParentTaskID = req.ParentTaskID
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) ListTasks(userID string, filters model.TaskFilters) ([]model.Task, error) {
	// Validate filters
	if filters.Status != "" && !isValidStatus(filters.Status) {
		return nil, ErrInvalidStatus
	}
	if filters.Priority != "" && !isValidPriority(filters.Priority) {
		return nil, ErrInvalidPriority
	}

	return s.taskRepo.List(userID, filters)
}

func (s *TaskService) GetTask(taskID, userID string) (*model.Task, error) {
	return s.taskRepo.GetByID(taskID, userID)
}

func (s *TaskService) UpdateTask(taskID, userID string, req *model.UpdateTaskRequest) (*model.Task, error) {
	// Get existing task
	task, err := s.taskRepo.GetByIDSimple(taskID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		if !isValidStatus(*req.Status) {
			return nil, ErrInvalidStatus
		}
		task.Status = *req.Status
	}
	if req.Priority != nil {
		if !isValidPriority(*req.Priority) {
			return nil, ErrInvalidPriority
		}
		task.Priority = *req.Priority
	}
	if req.Category != nil {
		task.Category = *req.Category
	}
	if req.EstimatedMinutes != nil {
		task.EstimatedMinutes = req.EstimatedMinutes
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			task.DueDate = nil
		} else {
			dueDate, err := time.Parse("2006-01-02", *req.DueDate)
			if err != nil {
				return nil, errors.New("invalid due date format: use YYYY-MM-DD")
			}
			task.DueDate = &dueDate
		}
	}

	if err := s.taskRepo.Update(task); err != nil {
		return nil, err
	}

	// Fetch updated task with subtasks
	return s.taskRepo.GetByID(taskID, userID)
}

func (s *TaskService) DeleteTask(taskID, userID string) error {
	return s.taskRepo.Delete(taskID, userID)
}

func (s *TaskService) GetCategories(userID string) ([]string, error) {
	return s.taskRepo.GetCategories(userID)
}

// CreateTaskFromAI creates a task from AI-generated data
func (s *TaskService) CreateTaskFromAI(userID string, generated model.GeneratedTask, parentID *string) (*model.Task, error) {
	task := &model.Task{
		UserID:       userID,
		ParentTaskID: parentID,
		Title:        generated.Title,
		Description:  generated.Description,
		Status:       "todo",
		Priority:     generated.Priority,
		Category:     generated.Category,
		AIGenerated:  true,
	}

	if generated.EstimatedMinutes > 0 {
		task.EstimatedMinutes = &generated.EstimatedMinutes
	}

	// Validate and fix priority
	if !isValidPriority(task.Priority) {
		task.Priority = "medium"
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, err
	}

	return task, nil
}

// BulkCreateFromAI creates multiple tasks from AI-generated data
func (s *TaskService) BulkCreateFromAI(userID string, generatedTasks []model.GeneratedTask) ([]model.Task, error) {
	var createdTasks []model.Task

	for _, gen := range generatedTasks {
		// Create parent task
		task, err := s.CreateTaskFromAI(userID, gen, nil)
		if err != nil {
			continue // Skip failed tasks
		}
		createdTasks = append(createdTasks, *task)

		// Create subtasks if any
		for _, subGen := range gen.Subtasks {
			subTask, err := s.CreateTaskFromAI(userID, subGen, &task.ID)
			if err != nil {
				continue
			}
			task.Subtasks = append(task.Subtasks, *subTask)
		}
	}

	return createdTasks, nil
}

// Helper functions
func isValidStatus(status string) bool {
	switch status {
	case "todo", "in_progress", "done":
		return true
	default:
		return false
	}
}

func isValidPriority(priority string) bool {
	switch priority {
	case "low", "medium", "high", "urgent":
		return true
	default:
		return false
	}
}
