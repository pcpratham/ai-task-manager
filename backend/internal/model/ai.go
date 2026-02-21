package model

// Request types for AI endpoints
type GenerateTasksRequest struct {
	Input string `json:"input" binding:"required,min=3"`
}

type SuggestPriorityRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type EstimateTimeRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

// Response types from Claude API parsing
type GeneratedTask struct {
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	Priority         string          `json:"priority"`
	Category         string          `json:"category"`
	EstimatedMinutes int             `json:"estimated_minutes"`
	Subtasks         []GeneratedTask `json:"subtasks,omitempty"`
}

type PrioritySuggestion struct {
	Priority string `json:"priority"`
	Reason   string `json:"reason"`
}

type TimeEstimate struct {
	EstimatedMinutes int    `json:"estimated_minutes"`
	Reasoning        string `json:"reasoning"`
}

// API Response wrappers
type GenerateTasksResponse struct {
	Tasks   []Task `json:"tasks"`
	Message string `json:"message"`
}

type BreakdownResponse struct {
	Subtasks []Task `json:"subtasks"`
	Message  string `json:"message"`
}

// Claude API types
type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
}

type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeResponse struct {
	Content []ClaudeContent `json:"content"`
	Error   *ClaudeError    `json:"error,omitempty"`
}

type ClaudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ClaudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
