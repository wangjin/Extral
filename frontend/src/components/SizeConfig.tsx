import { Task } from '../hooks/useTaskQueue'

interface Props {
  task: Task
  onChange: (task: Task) => void
}

export default function SizeConfig({ task, onChange }: Props) {
  const update = (patch: Partial<Task>) => {
    onChange({ ...task, ...patch })
  }

  return (
    <div className="config-section">
      <label className="config-label">输出尺寸</label>
      <div className="format-row">
        <div className="btn-group">
          <button
            className={`btn-option ${task.scaleMode === 'percentage' ? 'active' : ''}`}
            onClick={() => update({ scaleMode: 'percentage' })}
          >按比例</button>
          <button
            className={`btn-option ${task.scaleMode === 'custom' ? 'active' : ''}`}
            onClick={() => update({ scaleMode: 'custom' })}
          >自定义</button>
        </div>
        {task.scaleMode === 'percentage' && (
          <div className="param-row" style={{ marginLeft: 16 }}>
            <span className="param-label">缩放</span>
            <input
              type="range"
              min={10}
              max={200}
              value={task.scale}
              onChange={e => update({ scale: parseInt(e.target.value) })}
            />
            <span className="param-value">{task.scale}%</span>
          </div>
        )}
        {task.scaleMode === 'custom' && (
          <div className="param-row" style={{ marginLeft: 16 }}>
            <input
              type="number"
              className="input-sm"
              placeholder="宽"
              value={task.width || ''}
              onChange={e => update({ width: parseInt(e.target.value) || 0 })}
            />
            <span className="param-unit">×</span>
            <input
              type="number"
              className="input-sm"
              placeholder="高"
              value={task.height || ''}
              onChange={e => update({ height: parseInt(e.target.value) || 0 })}
            />
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={task.keepRatio}
                onChange={e => update({ keepRatio: e.target.checked })}
              />
              保持比例
            </label>
          </div>
        )}
      </div>
    </div>
  )
}
