import {
  AuthResponse,
  Task,
  TaskFilters,
  CreateTaskInput,
  UpdateTaskInput,
  PrioritySuggestion,
  TimeEstimate,
  GenerateTasksResponse,
  BreakdownResponse,
} from './types';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.status = status;
    this.name = 'ApiError';
  }
}

async function fetchAPI<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;

  const res = await fetch(`${API_URL}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  });

  if (res.status === 401) {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    throw new ApiError('Unauthorized', 401);
  }

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: 'Something went wrong' }));
    throw new ApiError(error.error || 'Something went wrong', res.status);
  }

  return res.json();
}

export const api = {
  // Auth
  register: (data: { email: string; password: string; name: string }) =>
    fetchAPI<AuthResponse>('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  login: (data: { email: string; password: string }) =>
    fetchAPI<AuthResponse>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  me: () => fetchAPI<{ user: { id: string; email: string; name: string } }>('/api/auth/me'),

  // Tasks
  getTasks: (filters?: TaskFilters) => {
    const params = new URLSearchParams();
    if (filters?.status) params.append('status', filters.status);
    if (filters?.priority) params.append('priority', filters.priority);
    if (filters?.category) params.append('category', filters.category);
    if (filters?.search) params.append('search', filters.search);

    const queryString = params.toString();
    return fetchAPI<{ tasks: Task[] }>(`/api/tasks${queryString ? `?${queryString}` : ''}`);
  },

  getTask: (id: string) => fetchAPI<Task>(`/api/tasks/${id}`),

  createTask: (data: CreateTaskInput) =>
    fetchAPI<Task>('/api/tasks', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  updateTask: (id: string, data: UpdateTaskInput) =>
    fetchAPI<Task>(`/api/tasks/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  deleteTask: (id: string) =>
    fetchAPI<{ message: string }>(`/api/tasks/${id}`, {
      method: 'DELETE',
    }),

  getCategories: () => fetchAPI<{ categories: string[] }>('/api/tasks/categories'),

  // AI
  aiGenerate: (input: string) =>
    fetchAPI<GenerateTasksResponse>('/api/ai/generate', {
      method: 'POST',
      body: JSON.stringify({ input }),
    }),

  aiBreakdown: (taskId: string) =>
    fetchAPI<BreakdownResponse>(`/api/ai/breakdown/${taskId}`, {
      method: 'POST',
    }),

  aiSuggestPriority: (data: { title: string; description?: string }) =>
    fetchAPI<PrioritySuggestion>('/api/ai/suggest-priority', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  aiEstimateTime: (data: { title: string; description?: string }) =>
    fetchAPI<TimeEstimate>('/api/ai/estimate-time', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
};

export { ApiError };
