'use client';

import { useState } from 'react';
import { Task, PRIORITY_COLORS, STATUS_COLORS, TaskPriority, TaskStatus } from '@/lib/types';
import { api } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { Card, CardContent } from '@/components/ui/card';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { toast } from 'sonner';
import {
  MoreHorizontal,
  Pencil,
  Trash2,
  Sparkles,
  Clock,
  Calendar,
  ChevronDown,
  ChevronRight,
  Loader2,
  Layers,
} from 'lucide-react';
import { format } from 'date-fns';

interface TaskCardProps {
  task: Task;
  onUpdate: (task: Task) => void;
  onDelete: (taskId: string) => void;
  onEdit: (task: Task) => void;
  onSubtasksGenerated: (parentId: string, subtasks: Task[]) => void;
}

export function TaskCard({ task, onUpdate, onDelete, onEdit, onSubtasksGenerated }: TaskCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [isBreakingDown, setIsBreakingDown] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  const handleStatusToggle = async () => {
    const newStatus: TaskStatus = task.status === 'done' ? 'todo' : 'done';

    try {
      const updatedTask = await api.updateTask(task.id, { status: newStatus });
      onUpdate(updatedTask);
      toast.success(newStatus === 'done' ? 'Task completed!' : 'Task reopened');
    } catch (error) {
      toast.error('Failed to update task');
    }
  };

  const handleBreakdown = async () => {
    setIsBreakingDown(true);

    try {
      const response = await api.aiBreakdown(task.id);
      onSubtasksGenerated(task.id, response.subtasks);
      toast.success(`Created ${response.subtasks.length} subtasks`);
      setIsExpanded(true);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to breakdown task');
    } finally {
      setIsBreakingDown(false);
    }
  };

  const handleDelete = async () => {
    setIsDeleting(true);

    try {
      await api.deleteTask(task.id);
      onDelete(task.id);
      toast.success('Task deleted');
    } catch (error) {
      toast.error('Failed to delete task');
    } finally {
      setIsDeleting(false);
    }
  };

  const hasSubtasks = task.subtasks && task.subtasks.length > 0;
  const subtaskCount = task.subtask_count || (task.subtasks?.length ?? 0);

  return (
    <Card className={`transition-all ${task.status === 'done' ? 'opacity-60' : ''}`}>
      <CardContent className="p-4">
        <div className="flex items-start gap-3">
          {/* Checkbox */}
          <Checkbox
            checked={task.status === 'done'}
            onCheckedChange={handleStatusToggle}
            className="mt-1"
          />

          {/* Main content */}
          <div className="flex-1 min-w-0">
            <div className="flex items-start justify-between gap-2">
              <div className="flex-1">
                {/* Title row */}
                <div className="flex items-center gap-2 flex-wrap">
                  <h3 className={`font-medium ${task.status === 'done' ? 'line-through text-muted-foreground' : ''}`}>
                    {task.title}
                  </h3>
                  {task.ai_generated && (
                    <Badge variant="outline" className="text-xs gap-1">
                      <Sparkles className="h-3 w-3" />
                      AI
                    </Badge>
                  )}
                </div>

                {/* Description */}
                {task.description && (
                  <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                    {task.description}
                  </p>
                )}

                {/* Badges row */}
                <div className="flex items-center gap-2 mt-2 flex-wrap">
                  <Badge className={PRIORITY_COLORS[task.priority as TaskPriority]}>
                    {task.priority}
                  </Badge>
                  <Badge variant="outline" className={STATUS_COLORS[task.status as TaskStatus]}>
                    {task.status.replace('_', ' ')}
                  </Badge>
                  {task.category && (
                    <Badge variant="secondary">{task.category}</Badge>
                  )}
                </div>

                {/* Meta info */}
                <div className="flex items-center gap-4 mt-2 text-xs text-muted-foreground">
                  {task.estimated_minutes && (
                    <span className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      {task.estimated_minutes} min
                    </span>
                  )}
                  {task.due_date && (
                    <span className="flex items-center gap-1">
                      <Calendar className="h-3 w-3" />
                      {format(new Date(task.due_date), 'MMM d, yyyy')}
                    </span>
                  )}
                  {subtaskCount > 0 && (
                    <button
                      onClick={() => setIsExpanded(!isExpanded)}
                      className="flex items-center gap-1 hover:text-foreground transition-colors"
                    >
                      <Layers className="h-3 w-3" />
                      {subtaskCount} subtasks
                      {isExpanded ? (
                        <ChevronDown className="h-3 w-3" />
                      ) : (
                        <ChevronRight className="h-3 w-3" />
                      )}
                    </button>
                  )}
                </div>
              </div>

              {/* Actions */}
              <div className="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  onClick={handleBreakdown}
                  disabled={isBreakingDown || task.status === 'done'}
                  title="Break down with AI"
                >
                  {isBreakingDown ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Sparkles className="h-4 w-4" />
                  )}
                </Button>

                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon" className="h-8 w-8">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => onEdit(task)}>
                      <Pencil className="mr-2 h-4 w-4" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      onClick={handleDelete}
                      disabled={isDeleting}
                      className="text-destructive focus:text-destructive"
                    >
                      {isDeleting ? (
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      ) : (
                        <Trash2 className="mr-2 h-4 w-4" />
                      )}
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </div>

            {/* Subtasks */}
            {isExpanded && hasSubtasks && (
              <div className="mt-4 pl-4 border-l-2 space-y-2">
                {task.subtasks!.map((subtask) => (
                  <SubtaskItem
                    key={subtask.id}
                    subtask={subtask}
                    onUpdate={onUpdate}
                  />
                ))}
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function SubtaskItem({ subtask, onUpdate }: { subtask: Task; onUpdate: (task: Task) => void }) {
  const handleStatusToggle = async () => {
    const newStatus: TaskStatus = subtask.status === 'done' ? 'todo' : 'done';

    try {
      const updatedTask = await api.updateTask(subtask.id, { status: newStatus });
      onUpdate(updatedTask);
    } catch (error) {
      toast.error('Failed to update subtask');
    }
  };

  return (
    <div className="flex items-center gap-2 py-1">
      <Checkbox
        checked={subtask.status === 'done'}
        onCheckedChange={handleStatusToggle}
        className="h-4 w-4"
      />
      <span className={`text-sm ${subtask.status === 'done' ? 'line-through text-muted-foreground' : ''}`}>
        {subtask.title}
      </span>
      {subtask.estimated_minutes && (
        <span className="text-xs text-muted-foreground">
          ({subtask.estimated_minutes} min)
        </span>
      )}
    </div>
  );
}
