export interface User {
  id: string;
  email: string;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface Task {
  id: string;
  user_id: string;
  parent_task_id?: string;
  title: string;
  description: string;
  status: 'todo' | 'in_progress' | 'done';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  category: string;
  estimated_minutes?: number;
  due_date?: string;
  ai_generated: boolean;
  created_at: string;
  updated_at: string;
  subtask_count?: number;
  subtasks?: Task[];
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface TaskFilters {
  status?: string;
  priority?: string;
  category?: string;
  search?: string;
}

export interface CreateTaskInput {
  title: string;
  description?: string;
  priority?: string;
  category?: string;
  estimated_minutes?: number;
  due_date?: string;
  parent_task_id?: string;
}

export interface UpdateTaskInput {
  title?: string;
  description?: string;
  status?: string;
  priority?: string;
  category?: string;
  estimated_minutes?: number;
  due_date?: string;
}

export interface PrioritySuggestion {
  priority: string;
  reason: string;
}

export interface TimeEstimate {
  estimated_minutes: number;
  reasoning: string;
}

export interface GenerateTasksResponse {
  tasks: Task[];
  message: string;
}

export interface BreakdownResponse {
  subtasks: Task[];
  message: string;
}

export type TaskStatus = 'todo' | 'in_progress' | 'done';
export type TaskPriority = 'low' | 'medium' | 'high' | 'urgent';

export const STATUS_LABELS: Record<TaskStatus, string> = {
  todo: 'To Do',
  in_progress: 'In Progress',
  done: 'Done',
};

export const PRIORITY_LABELS: Record<TaskPriority, string> = {
  low: 'Low',
  medium: 'Medium',
  high: 'High',
  urgent: 'Urgent',
};

export const PRIORITY_COLORS: Record<TaskPriority, string> = {
  low: 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300',
  medium: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
  high: 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300',
  urgent: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
};

export const STATUS_COLORS: Record<TaskStatus, string> = {
  todo: 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300',
  in_progress: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
  done: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
};
