package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/task-manager/backend/internal/model"
)

const (
	geminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"
)

var (
	ErrAIUnavailable = errors.New("AI service is currently unavailable")
	ErrAIResponse    = errors.New("failed to parse AI response")
)

// Gemini API types
type GeminiRequest struct {
	Contents         []GeminiContent        `json:"contents"`
	GenerationConfig GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
	Error      *GeminiError      `json:"error,omitempty"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type AIService struct {
	apiKey     string
	httpClient *http.Client
}

func NewAIService(apiKey string) *AIService {
	return &AIService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (s *AIService) call(prompt string) (string, error) {
	if s.apiKey == "" {
		log.Println("[AI] Error: API key is empty")
		return "", ErrAIUnavailable
	}

	log.Printf("[AI] Making request to Gemini API (prompt length: %d chars)", len(prompt))

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: GeminiGenerationConfig{
			Temperature:     0.7,
			MaxOutputTokens: 4096,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[AI] Error marshaling request: %v", err)
		return "", err
	}

	url := fmt.Sprintf("%s?key=%s", geminiAPIURL, s.apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("[AI] Error creating request: %v", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("[AI] Error making request: %v", err)
		return "", ErrAIUnavailable
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[AI] Error reading response body: %v", err)
		return "", err
	}

	log.Printf("[AI] Response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		log.Printf("[AI] API error response: %s", string(body))
		return "", fmt.Errorf("gemini API error: %s", string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		log.Printf("[AI] Error parsing response JSON: %v", err)
		return "", err
	}

	if geminiResp.Error != nil {
		log.Printf("[AI] Gemini returned error: %s", geminiResp.Error.Message)
		return "", fmt.Errorf("gemini API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		log.Println("[AI] Error: No candidates or parts in response")
		return "", ErrAIResponse
	}

	result := geminiResp.Candidates[0].Content.Parts[0].Text
	log.Printf("[AI] Success: Got response (%d chars)", len(result))

	return result, nil
}

func (s *AIService) GenerateTasks(input string) ([]model.GeneratedTask, error) {
	log.Printf("[AI] GenerateTasks called with input: %s", input)

	prompt := `You are a task planning assistant. Given a natural language request, break it into actionable, concrete tasks.

Rules:
- Create 3-8 top-level tasks
- Each task can optionally have 2-5 subtasks for complex items
- Assign realistic priorities: low, medium, high, or urgent
- Assign a category to group related tasks (e.g., "Planning", "Development", "Communication", "Research")
- Estimate time in minutes (be realistic: 15, 30, 60, 120, etc.)
- Titles should be action-oriented (start with a verb like "Create", "Review", "Send", "Prepare")
- Descriptions should be 1-2 sentences explaining the task

Return ONLY valid JSON with no markdown formatting, code blocks, or extra text. The response must be a JSON array:

[
  {
    "title": "Task title here",
    "description": "Brief description of what needs to be done",
    "priority": "medium",
    "category": "Planning",
    "estimated_minutes": 30,
    "subtasks": [
      {
        "title": "Subtask title",
        "description": "Subtask description",
        "priority": "medium",
        "category": "Planning",
        "estimated_minutes": 15
      }
    ]
  }
]

User request: ` + input

	response, err := s.call(prompt)
	if err != nil {
		return nil, err
	}

	return s.parseGeneratedTasks(response)
}

func (s *AIService) BreakdownTask(title, description string) ([]model.GeneratedTask, error) {
	prompt := fmt.Sprintf(`You are a task breakdown assistant. Given a task title and description, break it into 3-7 concrete, actionable subtasks.

Rules:
- Create specific, actionable subtasks that together complete the parent task
- Each subtask should be completable in one sitting
- Assign priorities relative to the parent task: low, medium, high, or urgent
- Keep the same category as the parent task or use a relevant sub-category
- Estimate time in minutes (15, 30, 60, 120, etc.)
- Subtasks should be ordered logically (dependencies first)

Return ONLY valid JSON with no markdown formatting, code blocks, or extra text. The response must be a JSON array:

[
  {
    "title": "Subtask title",
    "description": "What needs to be done",
    "priority": "medium",
    "category": "Same as parent",
    "estimated_minutes": 30
  }
]

Task: %s
Description: %s`, title, description)

	response, err := s.call(prompt)
	if err != nil {
		return nil, err
	}

	return s.parseGeneratedTasks(response)
}

func (s *AIService) SuggestPriority(title, description string) (*model.PrioritySuggestion, error) {
	prompt := fmt.Sprintf(`You are a task prioritization assistant. Given a task title and optional description, suggest an appropriate priority level.

Priority levels:
- urgent: Time-sensitive, blocking other work, or has immediate consequences if delayed
- high: Important for goals, has a deadline soon, or affects others
- medium: Standard tasks that should be done but aren't time-critical
- low: Nice-to-have, can be deferred, or minimal impact if delayed

Return ONLY valid JSON with no markdown formatting:

{
  "priority": "medium",
  "reason": "Brief explanation of why this priority was chosen"
}

Task: %s
Description: %s`, title, description)

	response, err := s.call(prompt)
	if err != nil {
		return nil, err
	}

	return s.parsePrioritySuggestion(response)
}

func (s *AIService) EstimateTime(title, description string) (*model.TimeEstimate, error) {
	prompt := fmt.Sprintf(`You are a project estimation assistant. Given a task title and description, estimate how long it would take a competent professional to complete.

Guidelines:
- Be realistic, not optimistic
- Consider preparation, execution, and review time
- Round to practical increments: 15, 30, 45, 60, 90, 120, 180, 240 minutes
- For very large tasks, suggest breaking them down

Return ONLY valid JSON with no markdown formatting:

{
  "estimated_minutes": 60,
  "reasoning": "Brief explanation of the estimate breakdown"
}

Task: %s
Description: %s`, title, description)

	response, err := s.call(prompt)
	if err != nil {
		return nil, err
	}

	return s.parseTimeEstimate(response)
}

// Parsing helpers

func (s *AIService) parseGeneratedTasks(response string) ([]model.GeneratedTask, error) {
	log.Printf("[AI] Parsing response (raw length: %d chars)", len(response))

	// Clean up response - remove markdown code blocks if present
	response = cleanJSONResponse(response)
	log.Printf("[AI] After cleanup (length: %d chars)", len(response))

	var tasks []model.GeneratedTask
	if err := json.Unmarshal([]byte(response), &tasks); err != nil {
		log.Printf("[AI] Initial parse failed: %v, trying to extract JSON array", err)
		// Try to find JSON array in response
		start := strings.Index(response, "[")
		end := strings.LastIndex(response, "]")
		if start != -1 && end != -1 && end > start {
			jsonStr := response[start : end+1]
			log.Printf("[AI] Extracted JSON substring (length: %d)", len(jsonStr))
			if err := json.Unmarshal([]byte(jsonStr), &tasks); err != nil {
				log.Printf("[AI] Parse failed after extraction: %v", err)
				log.Printf("[AI] Response content: %s", response[:min(500, len(response))])
				return nil, fmt.Errorf("%w: %v", ErrAIResponse, err)
			}
		} else {
			log.Printf("[AI] Could not find JSON array in response")
			log.Printf("[AI] Response content: %s", response[:min(500, len(response))])
			return nil, fmt.Errorf("%w: %v", ErrAIResponse, err)
		}
	}

	log.Printf("[AI] Successfully parsed %d tasks", len(tasks))

	// Validate and fix tasks
	for i := range tasks {
		if tasks[i].Priority == "" || !isValidAIPriority(tasks[i].Priority) {
			tasks[i].Priority = "medium"
		}
		if tasks[i].EstimatedMinutes <= 0 {
			tasks[i].EstimatedMinutes = 30
		}

		// Fix subtasks too
		for j := range tasks[i].Subtasks {
			if tasks[i].Subtasks[j].Priority == "" || !isValidAIPriority(tasks[i].Subtasks[j].Priority) {
				tasks[i].Subtasks[j].Priority = "medium"
			}
			if tasks[i].Subtasks[j].EstimatedMinutes <= 0 {
				tasks[i].Subtasks[j].EstimatedMinutes = 15
			}
		}
	}

	return tasks, nil
}

func (s *AIService) parsePrioritySuggestion(response string) (*model.PrioritySuggestion, error) {
	response = cleanJSONResponse(response)

	var suggestion model.PrioritySuggestion
	if err := json.Unmarshal([]byte(response), &suggestion); err != nil {
		// Try to find JSON object in response
		start := strings.Index(response, "{")
		end := strings.LastIndex(response, "}")
		if start != -1 && end != -1 && end > start {
			if err := json.Unmarshal([]byte(response[start:end+1]), &suggestion); err != nil {
				return nil, fmt.Errorf("%w: %v", ErrAIResponse, err)
			}
		} else {
			return nil, fmt.Errorf("%w: %v", ErrAIResponse, err)
		}
	}

	// Validate priority
	if !isValidAIPriority(suggestion.Priority) {
		suggestion.Priority = "medium"
	}

	return &suggestion, nil
}

func (s *AIService) parseTimeEstimate(response string) (*model.TimeEstimate, error) {
	response = cleanJSONResponse(response)

	var estimate model.TimeEstimate
	if err := json.Unmarshal([]byte(response), &estimate); err != nil {
		// Try to find JSON object in response
		start := strings.Index(response, "{")
		end := strings.LastIndex(response, "}")
		if start != -1 && end != -1 && end > start {
			if err := json.Unmarshal([]byte(response[start:end+1]), &estimate); err != nil {
				return nil, fmt.Errorf("%w: %v", ErrAIResponse, err)
			}
		} else {
			return nil, fmt.Errorf("%w: %v", ErrAIResponse, err)
		}
	}

	// Validate estimate
	if estimate.EstimatedMinutes <= 0 {
		estimate.EstimatedMinutes = 30
	}

	return &estimate, nil
}

func cleanJSONResponse(response string) string {
	// Remove markdown code blocks
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	return strings.TrimSpace(response)
}

func isValidAIPriority(priority string) bool {
	switch priority {
	case "low", "medium", "high", "urgent":
		return true
	default:
		return false
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
