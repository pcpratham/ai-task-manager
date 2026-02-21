package model

import "time"

type Task struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	ParentTaskID     *string    `json:"parent_task_id,omitempty"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	Status           string     `json:"status"`
	Priority         string     `json:"priority"`
	Category         string     `json:"category"`
	EstimatedMinutes *int       `json:"estimated_minutes,omitempty"`
	DueDate          *time.Time `json:"due_date,omitempty"`
	AIGenerated      bool       `json:"ai_generated"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	SubtaskCount     int        `json:"subtask_count,omitempty"`
	Subtasks         []Task     `json:"subtasks,omitempty"`
}

type CreateTaskRequest struct {
	Title            string  `json:"title" binding:"required,min=1"`
	Description      string  `json:"description"`
	Priority         string  `json:"priority"`
	Category         string  `json:"category"`
	EstimatedMinutes *int    `json:"estimated_minutes"`
	DueDate          *string `json:"due_date"`
	ParentTaskID     *string `json:"parent_task_id"`
}

type UpdateTaskRequest struct {
	Title            *string `json:"title"`
	Description      *string `json:"description"`
	Status           *string `json:"status"`
	Priority         *string `json:"priority"`
	Category         *string `json:"category"`
	EstimatedMinutes *int    `json:"estimated_minutes"`
	DueDate          *string `json:"due_date"`
}

type TaskFilters struct {
	Status   string `form:"status"`
	Priority string `form:"priority"`
	Category string `form:"category"`
	Search   string `form:"search"`
}

type TaskListResponse struct {
	Tasks []Task `json:"tasks"`
}
