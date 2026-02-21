package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/task-manager/backend/internal/model"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(task *model.Task) error {
	task.ID = uuid.New().String()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	if task.Status == "" {
		task.Status = "todo"
	}
	if task.Priority == "" {
		task.Priority = "medium"
	}

	query := `
		INSERT INTO tasks (id, user_id, parent_task_id, title, description, status, priority, category, estimated_minutes, due_date, ai_generated, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		task.ID,
		task.UserID,
		task.ParentTaskID,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.Category,
		task.EstimatedMinutes,
		task.DueDate,
		task.AIGenerated,
		task.CreatedAt,
		task.UpdatedAt,
	)

	return err
}

func (r *TaskRepository) List(userID string, filters model.TaskFilters) ([]model.Task, error) {
	query := `
		SELECT id, user_id, parent_task_id, title, description, status, priority, category, estimated_minutes, due_date, ai_generated, created_at, updated_at
		FROM tasks
		WHERE user_id = ? AND parent_task_id IS NULL
	`
	args := []interface{}{userID}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}
	if filters.Priority != "" {
		query += " AND priority = ?"
		args = append(args, filters.Priority)
	}
	if filters.Category != "" {
		query += " AND category = ?"
		args = append(args, filters.Category)
	}
	if filters.Search != "" {
		query += " AND (title LIKE ? OR description LIKE ?)"
		searchTerm := "%" + filters.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []model.Task{}
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		// Get subtask count
		count, err := r.CountSubtasks(task.ID)
		if err == nil {
			task.SubtaskCount = count
		}

		tasks = append(tasks, *task)
	}

	return tasks, nil
}

func (r *TaskRepository) GetByID(taskID, userID string) (*model.Task, error) {
	query := `
		SELECT id, user_id, parent_task_id, title, description, status, priority, category, estimated_minutes, due_date, ai_generated, created_at, updated_at
		FROM tasks
		WHERE id = ? AND user_id = ?
	`

	row := r.db.QueryRow(query, taskID, userID)
	task, err := scanTaskRow(row)
	if err == sql.ErrNoRows {
		return nil, ErrTaskNotFound
	}
	if err != nil {
		return nil, err
	}

	// Get subtasks
	subtasks, err := r.GetSubtasks(taskID, userID)
	if err == nil {
		task.Subtasks = subtasks
		task.SubtaskCount = len(subtasks)
	}

	return task, nil
}

func (r *TaskRepository) Update(task *model.Task) error {
	task.UpdatedAt = time.Now()

	query := `
		UPDATE tasks
		SET title = ?, description = ?, status = ?, priority = ?, category = ?, estimated_minutes = ?, due_date = ?, updated_at = ?
		WHERE id = ? AND user_id = ?
	`

	result, err := r.db.Exec(query,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.Category,
		task.EstimatedMinutes,
		task.DueDate,
		task.UpdatedAt,
		task.ID,
		task.UserID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) Delete(taskID, userID string) error {
	query := `DELETE FROM tasks WHERE id = ? AND user_id = ?`

	result, err := r.db.Exec(query, taskID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) GetSubtasks(parentID, userID string) ([]model.Task, error) {
	query := `
		SELECT id, user_id, parent_task_id, title, description, status, priority, category, estimated_minutes, due_date, ai_generated, created_at, updated_at
		FROM tasks
		WHERE parent_task_id = ? AND user_id = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, parentID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subtasks := []model.Task{}
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		subtasks = append(subtasks, *task)
	}

	return subtasks, nil
}

func (r *TaskRepository) CountSubtasks(taskID string) (int, error) {
	query := `SELECT COUNT(*) FROM tasks WHERE parent_task_id = ?`

	var count int
	err := r.db.QueryRow(query, taskID).Scan(&count)
	return count, err
}

func (r *TaskRepository) GetCategories(userID string) ([]string, error) {
	query := `
		SELECT DISTINCT category
		FROM tasks
		WHERE user_id = ? AND category != ''
		ORDER BY category
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []string{}
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// Helper functions for scanning
func scanTask(rows *sql.Rows) (*model.Task, error) {
	task := &model.Task{}
	var parentTaskID sql.NullString
	var estimatedMinutes sql.NullInt64
	var dueDate sql.NullTime

	err := rows.Scan(
		&task.ID,
		&task.UserID,
		&parentTaskID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.Category,
		&estimatedMinutes,
		&dueDate,
		&task.AIGenerated,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if parentTaskID.Valid {
		task.ParentTaskID = &parentTaskID.String
	}
	if estimatedMinutes.Valid {
		mins := int(estimatedMinutes.Int64)
		task.EstimatedMinutes = &mins
	}
	if dueDate.Valid {
		task.DueDate = &dueDate.Time
	}

	return task, nil
}

func scanTaskRow(row *sql.Row) (*model.Task, error) {
	task := &model.Task{}
	var parentTaskID sql.NullString
	var estimatedMinutes sql.NullInt64
	var dueDate sql.NullTime

	err := row.Scan(
		&task.ID,
		&task.UserID,
		&parentTaskID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.Category,
		&estimatedMinutes,
		&dueDate,
		&task.AIGenerated,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if parentTaskID.Valid {
		task.ParentTaskID = &parentTaskID.String
	}
	if estimatedMinutes.Valid {
		mins := int(estimatedMinutes.Int64)
		task.EstimatedMinutes = &mins
	}
	if dueDate.Valid {
		task.DueDate = &dueDate.Time
	}

	return task, nil
}

// BulkCreate creates multiple tasks in a single transaction
func (r *TaskRepository) BulkCreate(tasks []model.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO tasks (id, user_id, parent_task_id, title, description, status, priority, category, estimated_minutes, due_date, ai_generated, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for i := range tasks {
		tasks[i].ID = uuid.New().String()
		tasks[i].CreatedAt = now
		tasks[i].UpdatedAt = now

		if tasks[i].Status == "" {
			tasks[i].Status = "todo"
		}
		if tasks[i].Priority == "" {
			tasks[i].Priority = "medium"
		}

		_, err := stmt.Exec(
			tasks[i].ID,
			tasks[i].UserID,
			tasks[i].ParentTaskID,
			tasks[i].Title,
			tasks[i].Description,
			tasks[i].Status,
			tasks[i].Priority,
			tasks[i].Category,
			tasks[i].EstimatedMinutes,
			tasks[i].DueDate,
			tasks[i].AIGenerated,
			tasks[i].CreatedAt,
			tasks[i].UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert task %s: %w", tasks[i].Title, err)
		}
	}

	return tx.Commit()
}

// GetByIDSimple returns a task without subtasks (for internal use)
func (r *TaskRepository) GetByIDSimple(taskID, userID string) (*model.Task, error) {
	query := `
		SELECT id, user_id, parent_task_id, title, description, status, priority, category, estimated_minutes, due_date, ai_generated, created_at, updated_at
		FROM tasks
		WHERE id = ? AND user_id = ?
	`

	row := r.db.QueryRow(query, taskID, userID)
	return scanTaskRow(row)
}

// Unused helper - keeping for reference
func buildFilterQuery(filters model.TaskFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if filters.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filters.Status)
	}
	if filters.Priority != "" {
		conditions = append(conditions, "priority = ?")
		args = append(args, filters.Priority)
	}
	if filters.Category != "" {
		conditions = append(conditions, "category = ?")
		args = append(args, filters.Category)
	}
	if filters.Search != "" {
		conditions = append(conditions, "(title LIKE ? OR description LIKE ?)")
		searchTerm := "%" + filters.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return " AND " + strings.Join(conditions, " AND "), args
}
