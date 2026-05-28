import { Task } from '../hooks/useTaskQueue'

interface Props {
  tasks: Task[]
  onStartAll: () => void
  onClear: () => void
}

export default function ActionBar({ tasks, onStartAll, onClear }: Props) {
  const pendingCount = tasks.filter(t => t.status === 'pending').length
  const totalFrames = tasks.reduce((sum, t) => sum + t.total, 0)
  const hasRunning = tasks.some(t => t.status === 'running')

  return (
    <footer className="action-bar">
      <div className="action-bar-info">
        已选择 {tasks.length} 个视频
        {totalFrames > 0 && ` · 预计输出 ${totalFrames.toLocaleString()} 帧`}
      </div>
      <div className="action-bar-buttons">
        <button
          className="btn btn-primary"
          disabled={pendingCount === 0 || hasRunning}
          onClick={onStartAll}
        >
          {hasRunning ? '处理中...' : `全部开始${pendingCount > 0 ? ` (${pendingCount})` : ''}`}
        </button>
        <button
          className="btn btn-secondary"
          disabled={tasks.length === 0 || hasRunning}
          onClick={onClear}
        >清空队列</button>
      </div>
    </footer>
  )
}
