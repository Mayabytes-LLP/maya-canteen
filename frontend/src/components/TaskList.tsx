import { Button } from "@/components/ui/button";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { Task, taskService } from "@/services/taskService";
import TaskForm from "@/components/TaskForm";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "./ui/card";

const TaskList = () => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [editingTask, setEditingTask] = useState<Task | null>(null);

  const fetchTasks = async () => {
    try {
      const tasks = await taskService.getAllTasks();
      setTasks(tasks);
    } catch (error) {
      toast.error("Failed to fetch tasks");
      console.error("Error fetching tasks:", error);
    }
  };

  useEffect(() => {
    fetchTasks();
  }, []);

  const handleCreateTask = async (
    task: Omit<Task, "id" | "created_at" | "updated_at">,
  ) => {
    try {
      await taskService.createTask(task);
      toast.success("Task created successfully");
      fetchTasks();
    } catch (error) {
      toast.error("Failed to create task");
      console.error("Error creating task:", error);
    }
  };

  const handleUpdateTask = async (id: number, task: Partial<Task>) => {
    try {
      await taskService.updateTask(id, task);
      toast.success("Task updated successfully");
      fetchTasks();
      setEditingTask(null);
    } catch (error) {
      toast.error("Failed to update task");
      console.error("Error updating task:", error);
    }
  };

  const handleDeleteTask = async (id: number) => {
    try {
      await taskService.deleteTask(id);
      toast.success("Task deleted successfully");
      fetchTasks();
    } catch (error) {
      toast.error("Failed to delete task");
      console.error("Error deleting task:", error);
    }
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-4xl font-bold mb-8">Task Manager</h1>

      <div className="mb-8">
        <h2 className="text-2xl font-semibold mb-4">Create New Task</h2>
        <div className="bg-card p-6 rounded-lg border shadow-sm">
          <TaskForm onSubmit={handleCreateTask} />
        </div>
      </div>

      <div>
        <h2 className="text-2xl font-semibold mb-4">Tasks</h2>
        <div className="grid gap-4">
          {tasks && tasks.length > 0 ? (
            tasks.map((task) => (
              <div
                key={task.id}
                className="bg-card p-6 rounded-lg border shadow-sm"
              >
                {editingTask?.id === task.id ? (
                  <TaskForm
                    task={task}
                    onSubmit={(updatedTask) =>
                      handleUpdateTask(task.id, updatedTask)
                    }
                    onCancel={() => setEditingTask(null)}
                  />
                ) : (
                  <Card>
                    <CardHeader>
                      <CardTitle>{task.title}</CardTitle>
                      <CardDescription>{task.description}</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <span
                        className={`inline-block px-2 rounded text-sm mt-2 ${
                          task.status === "completed"
                            ? "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100"
                            : "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100"
                        }`}
                      >
                        {task.status}
                      </span>
                    </CardContent>
                    <CardFooter>
                      <div className="space-x-2 tems-stretch">
                        <Button
                          variant="outline"
                          onClick={() => setEditingTask(task)}
                        >
                          Edit
                        </Button>
                        <Button
                          variant="destructive"
                          onClick={() => handleDeleteTask(task.id)}
                        >
                          Delete
                        </Button>
                      </div>
                    </CardFooter>
                  </Card>
                )}
              </div>
            ))
          ) : (
            <Card className="bg-slate-300">
              <CardContent>No tasks found</CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
};

export default TaskList;
