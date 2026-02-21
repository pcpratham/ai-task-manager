'use client';

import { useState, useEffect, useCallback } from 'react';
import { Task, TaskFilters } from '@/lib/types';
import { api } from '@/lib/api';
import { AIInput } from '@/components/AIInput';
import { TaskList } from '@/components/TaskList';
import { TaskForm } from '@/components/TaskForm';
import { FilterSidebar } from '@/components/FilterSidebar';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { toast } from 'sonner';
import { Plus, ListTodo, CheckCircle2, Clock } from 'lucide-react';

export default function DashboardPage() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [categories, setCategories] = useState<string[]>([]);
  const [filters, setFilters] = useState<TaskFilters>({});
  const [isLoading, setIsLoading] = useState(true);
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);

  const fetchTasks = useCallback(async () => {
    setIsLoading(true);
    try {
      const [tasksResponse, categoriesResponse] = await Promise.all([
        api.getTasks(filters),
        api.getCategories(),
      ]);
      setTasks(tasksResponse.tasks || []);
      setCategories(categoriesResponse.categories || []);
    } catch (error) {
      toast.error('Failed to load tasks');
    } finally {
      setIsLoading(false);
    }
  }, [filters]);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  const handleTasksGenerated = (newTasks: Task[]) => {
    setTasks((prev) => [...newTasks, ...prev]);
    // Refresh categories
    api.getCategories().then((res) => setCategories(res.categories || []));
  };

  const handleTaskUpdate = (updatedTask: Task) => {
    setTasks((prev) =>
      prev.map((t) => {
        if (t.id === updatedTask.id) {
          return updatedTask;
        }
        // Also check subtasks
        if (t.subtasks) {
          return {
            ...t,
            subtasks: t.subtasks.map((st) =>
              st.id === updatedTask.id ? updatedTask : st
            ),
          };
        }
        return t;
      })
    );
  };

  const handleTaskDelete = (taskId: string) => {
    setTasks((prev) => prev.filter((t) => t.id !== taskId));
  };

  const handleTaskEdit = (task: Task) => {
    setEditingTask(task);
    setIsFormOpen(true);
  };

  const handleSubtasksGenerated = (parentId: string, subtasks: Task[]) => {
    setTasks((prev) =>
      prev.map((t) => {
        if (t.id === parentId) {
          return {
            ...t,
            subtasks: [...(t.subtasks || []), ...subtasks],
            subtask_count: (t.subtask_count || 0) + subtasks.length,
          };
        }
        return t;
      })
    );
  };

  const handleFormSubmit = (task: Task) => {
    if (editingTask) {
      handleTaskUpdate(task);
    } else {
      setTasks((prev) => [task, ...prev]);
    }
    setEditingTask(null);
    // Refresh categories
    api.getCategories().then((res) => setCategories(res.categories || []));
  };

  const handleFormOpenChange = (open: boolean) => {
    setIsFormOpen(open);
    if (!open) {
      setEditingTask(null);
    }
  };

  // Stats
  const totalTasks = tasks.length;
  const completedTasks = tasks.filter((t) => t.status === 'done').length;
  const inProgressTasks = tasks.filter((t) => t.status === 'in_progress').length;

  return (
    <div className="space-y-6">
      {/* AI Input Section */}
      <Card>
        <CardContent className="p-6">
          <AIInput onTasksGenerated={handleTasksGenerated} />
        </CardContent>
      </Card>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardContent className="p-4 flex items-center gap-4">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900">
              <ListTodo className="h-5 w-5 text-blue-600 dark:text-blue-300" />
            </div>
            <div>
              <p className="text-2xl font-bold">{totalTasks}</p>
              <p className="text-sm text-muted-foreground">Total Tasks</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4 flex items-center gap-4">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-yellow-100 dark:bg-yellow-900">
              <Clock className="h-5 w-5 text-yellow-600 dark:text-yellow-300" />
            </div>
            <div>
              <p className="text-2xl font-bold">{inProgressTasks}</p>
              <p className="text-sm text-muted-foreground">In Progress</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4 flex items-center gap-4">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-green-100 dark:bg-green-900">
              <CheckCircle2 className="h-5 w-5 text-green-600 dark:text-green-300" />
            </div>
            <div>
              <p className="text-2xl font-bold">{completedTasks}</p>
              <p className="text-sm text-muted-foreground">Completed</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Main Content */}
      <div className="grid gap-6 lg:grid-cols-[280px_1fr]">
        {/* Sidebar */}
        <aside className="space-y-6">
          <Button onClick={() => setIsFormOpen(true)} className="w-full gap-2">
            <Plus className="h-4 w-4" />
            New Task
          </Button>
          <Separator />
          <FilterSidebar
            filters={filters}
            categories={categories}
            onFiltersChange={setFilters}
          />
        </aside>

        {/* Task List */}
        <div>
          <TaskList
            tasks={tasks}
            isLoading={isLoading}
            onUpdate={handleTaskUpdate}
            onDelete={handleTaskDelete}
            onEdit={handleTaskEdit}
            onSubtasksGenerated={handleSubtasksGenerated}
          />
        </div>
      </div>

      {/* Task Form Modal */}
      <TaskForm
        open={isFormOpen}
        onOpenChange={handleFormOpenChange}
        task={editingTask}
        onSubmit={handleFormSubmit}
      />
    </div>
  );
}
