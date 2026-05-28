import TaskItem from './TaskItem'
import { Task } from '../hooks/useTaskQueue'

interface Props {
  tasks: Task[]
  selectedTaskId: string | null
  onSelectTask: (id: string) => void
  onRemoveTask: (id: string) => void
  onCancelTask: (id: string) => void
}

export default function TaskList({ tasks, selectedTaskId, onSelectTask, onRemoveTask, onCancelTask }: Props) {
  if (tasks.length === 0) {
    return (
      <div className="task-list-empty">
        <p>点击上方"选择视频"或"选择文件夹"添加视频文件</p>
      </div>
    )
  }

  return (
    <div className="task-list">
      <div className="task-list-header">
        <span className="task-list-count">任务列表 ({tasks.length})</span>
      </div>
      <div className="task-list-items">
        {tasks.map(task => (
          <TaskItem
            key={task.id}
            task={task}
            selected={task.id === selectedTaskId}
            onSelect={() => onSelectTask(task.id)}
            onRemove={() => onRemoveTask(task.id)}
            onCancel={() => onCancelTask(task.id)}
          />
        ))}
      </div>
    </div>
  )
}
