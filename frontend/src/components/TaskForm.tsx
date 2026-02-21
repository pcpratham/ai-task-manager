'use client';

import { useState, useEffect } from 'react';
import { Task, CreateTaskInput, UpdateTaskInput, PRIORITY_LABELS, STATUS_LABELS } from '@/lib/types';
import { api } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { toast } from 'sonner';
import { Loader2, Sparkles, Clock, Lightbulb } from 'lucide-react';

interface TaskFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  task?: Task | null;
  onSubmit: (task: Task) => void;
}

export function TaskForm({ open, onOpenChange, task, onSubmit }: TaskFormProps) {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [priority, setPriority] = useState('medium');
  const [status, setStatus] = useState('todo');
  const [category, setCategory] = useState('');
  const [estimatedMinutes, setEstimatedMinutes] = useState<string>('');
  const [dueDate, setDueDate] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isSuggestingPriority, setIsSuggestingPriority] = useState(false);
  const [isEstimatingTime, setIsEstimatingTime] = useState(false);
  const [prioritySuggestion, setPrioritySuggestion] = useState<{ priority: string; reason: string } | null>(null);
  const [timeEstimate, setTimeEstimate] = useState<{ estimated_minutes: number; reasoning: string } | null>(null);

  const isEditing = !!task;

  useEffect(() => {
    if (task) {
      setTitle(task.title);
      setDescription(task.description || '');
      setPriority(task.priority);
      setStatus(task.status);
      setCategory(task.category || '');
      setEstimatedMinutes(task.estimated_minutes?.toString() || '');
      setDueDate(task.due_date ? task.due_date.split('T')[0] : '');
    } else {
      resetForm();
    }
  }, [task, open]);

  const resetForm = () => {
    setTitle('');
    setDescription('');
    setPriority('medium');
    setStatus('todo');
    setCategory('');
    setEstimatedMinutes('');
    setDueDate('');
    setPrioritySuggestion(null);
    setTimeEstimate(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!title.trim()) {
      toast.error('Please enter a task title');
      return;
    }

    setIsLoading(true);

    try {
      let result: Task;

      if (isEditing) {
        const updateData: UpdateTaskInput = {
          title,
          description,
          priority,
          status,
          category,
          estimated_minutes: estimatedMinutes ? parseInt(estimatedMinutes) : undefined,
          due_date: dueDate || undefined,
        };
        result = await api.updateTask(task!.id, updateData);
        toast.success('Task updated');
      } else {
        const createData: CreateTaskInput = {
          title,
          description,
          priority,
          category,
          estimated_minutes: estimatedMinutes ? parseInt(estimatedMinutes) : undefined,
          due_date: dueDate || undefined,
        };
        result = await api.createTask(createData);
        toast.success('Task created');
      }

      onSubmit(result);
      onOpenChange(false);
      resetForm();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to save task');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSuggestPriority = async () => {
    if (!title.trim()) {
      toast.error('Please enter a task title first');
      return;
    }

    setIsSuggestingPriority(true);
    setPrioritySuggestion(null);

    try {
      const suggestion = await api.aiSuggestPriority({ title, description });
      setPrioritySuggestion(suggestion);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to suggest priority');
    } finally {
      setIsSuggestingPriority(false);
    }
  };

  const handleEstimateTime = async () => {
    if (!title.trim()) {
      toast.error('Please enter a task title first');
      return;
    }

    setIsEstimatingTime(true);
    setTimeEstimate(null);

    try {
      const estimate = await api.aiEstimateTime({ title, description });
      setTimeEstimate(estimate);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to estimate time');
    } finally {
      setIsEstimatingTime(false);
    }
  };

  const acceptPrioritySuggestion = () => {
    if (prioritySuggestion) {
      setPriority(prioritySuggestion.priority);
      setPrioritySuggestion(null);
      toast.success('Priority applied');
    }
  };

  const acceptTimeEstimate = () => {
    if (timeEstimate) {
      setEstimatedMinutes(timeEstimate.estimated_minutes.toString());
      setTimeEstimate(null);
      toast.success('Time estimate applied');
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{isEditing ? 'Edit Task' : 'Create New Task'}</DialogTitle>
          <DialogDescription>
            {isEditing ? 'Update the task details below.' : 'Fill in the details for your new task.'}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Title */}
          <div className="space-y-2">
            <Label htmlFor="title">Title *</Label>
            <Input
              id="title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Enter task title"
              disabled={isLoading}
              required
            />
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Add more details about this task"
              disabled={isLoading}
              rows={3}
            />
          </div>

          {/* Priority and Status row */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Priority</Label>
              <Select value={priority} onValueChange={setPriority} disabled={isLoading}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {Object.entries(PRIORITY_LABELS).map(([value, label]) => (
                    <SelectItem key={value} value={value}>
                      {label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {isEditing && (
              <div className="space-y-2">
                <Label>Status</Label>
                <Select value={status} onValueChange={setStatus} disabled={isLoading}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {Object.entries(STATUS_LABELS).map(([value, label]) => (
                      <SelectItem key={value} value={value}>
                        {label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}
          </div>

          {/* AI Suggestions */}
          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleSuggestPriority}
              disabled={isSuggestingPriority || !title.trim()}
              className="gap-1"
            >
              {isSuggestingPriority ? (
                <Loader2 className="h-3 w-3 animate-spin" />
              ) : (
                <Lightbulb className="h-3 w-3" />
              )}
              Suggest Priority
            </Button>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleEstimateTime}
              disabled={isEstimatingTime || !title.trim()}
              className="gap-1"
            >
              {isEstimatingTime ? (
                <Loader2 className="h-3 w-3 animate-spin" />
              ) : (
                <Clock className="h-3 w-3" />
              )}
              Estimate Time
            </Button>
          </div>

          {/* Priority Suggestion */}
          {prioritySuggestion && (
            <div className="rounded-lg border bg-muted/50 p-3 space-y-2">
              <div className="flex items-center gap-2">
                <Sparkles className="h-4 w-4 text-primary" />
                <span className="font-medium text-sm">AI Suggestion: {prioritySuggestion.priority.toUpperCase()}</span>
              </div>
              <p className="text-sm text-muted-foreground">{prioritySuggestion.reason}</p>
              <div className="flex gap-2">
                <Button type="button" size="sm" onClick={acceptPrioritySuggestion}>
                  Accept
                </Button>
                <Button type="button" size="sm" variant="ghost" onClick={() => setPrioritySuggestion(null)}>
                  Dismiss
                </Button>
              </div>
            </div>
          )}

          {/* Time Estimate */}
          {timeEstimate && (
            <div className="rounded-lg border bg-muted/50 p-3 space-y-2">
              <div className="flex items-center gap-2">
                <Sparkles className="h-4 w-4 text-primary" />
                <span className="font-medium text-sm">AI Estimate: {timeEstimate.estimated_minutes} minutes</span>
              </div>
              <p className="text-sm text-muted-foreground">{timeEstimate.reasoning}</p>
              <div className="flex gap-2">
                <Button type="button" size="sm" onClick={acceptTimeEstimate}>
                  Accept
                </Button>
                <Button type="button" size="sm" variant="ghost" onClick={() => setTimeEstimate(null)}>
                  Dismiss
                </Button>
              </div>
            </div>
          )}

          {/* Category and Estimated Time */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="category">Category</Label>
              <Input
                id="category"
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                placeholder="e.g., Work, Personal"
                disabled={isLoading}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="estimatedMinutes">Est. Time (min)</Label>
              <Input
                id="estimatedMinutes"
                type="number"
                value={estimatedMinutes}
                onChange={(e) => setEstimatedMinutes(e.target.value)}
                placeholder="e.g., 30"
                min="1"
                disabled={isLoading}
              />
            </div>
          </div>

          {/* Due Date */}
          <div className="space-y-2">
            <Label htmlFor="dueDate">Due Date</Label>
            <Input
              id="dueDate"
              type="date"
              value={dueDate}
              onChange={(e) => setDueDate(e.target.value)}
              disabled={isLoading}
            />
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isLoading}>
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  {isEditing ? 'Updating...' : 'Creating...'}
                </>
              ) : (
                isEditing ? 'Update Task' : 'Create Task'
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
