'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { api } from '@/lib/api';
import { Task } from '@/lib/types';
import { toast } from 'sonner';
import { Sparkles, Loader2, Send } from 'lucide-react';

interface AIInputProps {
  onTasksGenerated: (tasks: Task[]) => void;
}

export function AIInput({ onTasksGenerated }: AIInputProps) {
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!input.trim()) {
      toast.error('Please enter a task description');
      return;
    }

    setIsLoading(true);

    try {
      const response = await api.aiGenerate(input);
      onTasksGenerated(response.tasks);
      setInput('');
      toast.success(`Generated ${response.tasks.length} tasks!`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to generate tasks');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="w-full">
      <div className="relative flex items-center gap-2">
        <div className="relative flex-1">
          <Sparkles className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            type="text"
            placeholder="Describe your tasks... e.g., 'Plan a product launch for next month'"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            disabled={isLoading}
            className="pl-10 pr-4 h-12 text-base"
          />
        </div>
        <Button
          type="submit"
          disabled={isLoading || !input.trim()}
          className="h-12 px-6"
        >
          {isLoading ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Generating...
            </>
          ) : (
            <>
              <Send className="mr-2 h-4 w-4" />
              Generate
            </>
          )}
        </Button>
      </div>
      <p className="mt-2 text-xs text-muted-foreground">
        AI will create organized tasks with priorities and time estimates
      </p>
    </form>
  );
}
