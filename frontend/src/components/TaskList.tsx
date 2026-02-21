'use client';

import { Task } from '@/lib/types';
import { TaskCard } from './TaskCard';
import { Loader2, Inbox } from 'lucide-react';

interface TaskListProps {
  tasks: Task[];
  isLoading: boolean;
  onUpdate: (task: Task) => void;
  onDelete: (taskId: string) => void;
  onEdit: (task: Task) => void;
  onSubtasksGenerated: (parentId: string, subtasks: Task[]) => void;
}

export function TaskList({
  tasks,
  isLoading,
  onUpdate,
  onDelete,
  onEdit,
  onSubtasksGenerated,
}: TaskListProps) {
  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (tasks.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-muted mb-4">
          <Inbox className="h-8 w-8 text-muted-foreground" />
        </div>
        <h3 className="font-semibold text-lg">No tasks yet</h3>
        <p className="text-muted-foreground mt-1 max-w-sm">
          Create your first task manually or use AI to generate tasks from a description.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {tasks.map((task) => (
        <TaskCard
          key={task.id}
          task={task}
          onUpdate={onUpdate}
          onDelete={onDelete}
          onEdit={onEdit}
          onSubtasksGenerated={onSubtasksGenerated}
        />
      ))}
    </div>
  );
}
