import { Task } from '../hooks/useTaskQueue'

interface Props {
  task: Task
  onChange: (task: Task) => void
}

export default function OutputFormat({ task, onChange }: Props) {
  const update = (patch: Partial<Task>) => {
    onChange({ ...task, ...patch })
  }

  return (
    <div className="config-section">
      <label className="config-label">输出格式</label>
      <div className="format-row">
        <div className="btn-group">
          <button
            className={`btn-option ${task.format === 'jpeg' ? 'active' : ''}`}
            onClick={() => update({ format: 'jpeg' })}
          >JPEG</button>
          <button
            className={`btn-option ${task.format === 'png' ? 'active' : ''}`}
            onClick={() => update({ format: 'png' })}
          >PNG</button>
        </div>
        {task.format === 'jpeg' && (
          <div className="param-row" style={{ marginLeft: 16 }}>
            <span className="param-label">质量</span>
            <input
              type="range"
              min={1}
              max={100}
              value={task.quality}
              onChange={e => update({ quality: parseInt(e.target.value) })}
            />
            <span className="param-value">{task.quality}%</span>
          </div>
        )}
      </div>
    </div>
  )
}
