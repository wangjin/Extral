import { Task } from '../hooks/useTaskQueue'

interface Props {
  task: Task
  onChange: (task: Task) => void
}

const NAMINGS = [
  { value: 'sequential', label: '序号' },
  { value: 'timestamp', label: '时间戳' },
  { value: 'video-seq', label: '视频名+序号' },
  { value: 'video-time', label: '视频名+时间' },
]

export default function NamingSelector({ task, onChange }: Props) {
  return (
    <div className="config-section">
      <label className="config-label">文件命名</label>
      <div className="btn-group">
        {NAMINGS.map(n => (
          <button
            key={n.value}
            className={`btn-option ${task.naming === n.value ? 'active' : ''}`}
            onClick={() => onChange({ ...task, naming: n.value })}
          >
            {n.label}
          </button>
        ))}
      </div>
    </div>
  )
}
