import { Task } from '../hooks/useTaskQueue'

interface Props {
  task: Task
  selected: boolean
  onSelect: () => void
  onRemove: () => void
  onCancel: () => void
}

function formatTime(seconds: number): string {
  if (seconds < 60) return `${seconds}秒`
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m}分${s}秒`
}

function getVideoName(path: string): string {
  return path.split('/').pop() || path.split('\\').pop() || path
}

const MODE_LABELS: Record<string, string> = {
  'every-frame': '每N帧',
  'every-second': '每N秒',
  'time-range': '时间范围',
  'total-frames': '总帧数',
  'keyframe': '关键帧',
  'scene-change': '场景切换',
  'timestamps': '指定时间点',
}

function StatusBadge({ status, progress, elapsedTime }: { status: string; progress: number; elapsedTime: number }) {
  switch (status) {
    case 'running':
      return (
        <div className="task-progress">
          <div className="progress-bar">
            <div className="progress-fill" style={{ width: `${progress}%` }} />
          </div>
          <span className="progress-text">
            {Math.round(progress)}% · 剩余 {formatTime(elapsedTime)}
          </span>
        </div>
      )
    case 'completed':
      return <span className="status status-completed">已完成</span>
    case 'failed':
      return <span className="status status-failed">失败</span>
    case 'cancelled':
      return <span className="status status-cancelled">已取消</span>
    default:
      return <span className="status status-pending">等待中</span>
  }
}

export default function TaskItem({ task, selected, onSelect, onRemove, onCancel }: Props) {
  const statusColor: Record<string, string> = {
    running: '#3b82f6',
    completed: '#22c55e',
    failed: '#ef4444',
    cancelled: '#94a3b8',
    pending: '#94a3b8',
  }

  return (
    <div
      className={`task-item ${selected ? 'task-item-selected' : ''}`}
      onClick={onSelect}
    >
      <div className="task-dot" style={{ background: statusColor[task.status] || '#94a3b8' }} />
      <div className="task-info">
        <div className="task-name">{getVideoName(task.videoPath)}</div>
        <div className="task-summary">
          {MODE_LABELS[task.mode] || task.mode} → {task.format.toUpperCase()}
          {task.scaleMode === 'percentage' && task.scale !== 100 && ` → 缩放${task.scale}%`}
          {task.scaleMode === 'custom' && task.width > 0 && ` → ${task.width}x${task.height}`}
        </div>
        {task.status === 'failed' && task.error && (
          <div className="task-error">{task.error}</div>
        )}
      </div>
      <div className="task-right">
        <StatusBadge status={task.status} progress={task.progress} elapsedTime={task.elapsedTime} />
        <div className="task-actions">
          <button className="btn-icon" onClick={(e) => { e.stopPropagation(); task.status === 'running' ? onCancel() : onRemove() }} title={task.status === 'running' ? '取消' : '移除'}>✕</button>
        </div>
      </div>
    </div>
  )
}
