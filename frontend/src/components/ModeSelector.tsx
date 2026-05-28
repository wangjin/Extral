import { Task } from '../hooks/useTaskQueue'

interface Props {
  task: Task
  onChange: (task: Task) => void
}

const MODES = [
  { value: 'every-frame', label: '每N帧' },
  { value: 'every-second', label: '每N秒' },
  { value: 'time-range', label: '时间范围' },
  { value: 'total-frames', label: '总帧数' },
  { value: 'keyframe', label: '关键帧' },
  { value: 'scene-change', label: '场景切换' },
  { value: 'timestamps', label: '指定时间点' },
]

export default function ModeSelector({ task, onChange }: Props) {
  const update = (patch: Partial<Task>) => {
    onChange({ ...task, ...patch })
  }

  return (
    <div className="config-section">
      <label className="config-label">抽帧模式</label>
      <div className="btn-group">
        {MODES.map(m => (
          <button
            key={m.value}
            className={`btn-option ${task.mode === m.value ? 'active' : ''}`}
            onClick={() => update({ mode: m.value })}
          >
            {m.label}
          </button>
        ))}
      </div>
      <div className="config-params">
        {task.mode === 'every-frame' && (
          <div className="param-row">
            <span className="param-label">间隔</span>
            <input
              type="number"
              className="input-sm"
              value={task.params.frameInterval || 1}
              min={1}
              onChange={e => update({
                params: { ...task.params, frameInterval: parseInt(e.target.value) || 1 }
              })}
            />
            <span className="param-unit">帧</span>
          </div>
        )}
        {task.mode === 'every-second' && (
          <div className="param-row">
            <span className="param-label">间隔</span>
            <input
              type="number"
              className="input-sm"
              value={task.params.secondInterval || 1}
              min={0.1}
              step={0.5}
              onChange={e => update({
                params: { ...task.params, secondInterval: parseFloat(e.target.value) || 0.5 }
              })}
            />
            <span className="param-unit">秒</span>
          </div>
        )}
        {task.mode === 'time-range' && (
          <div className="param-row">
            <input
              type="text"
              className="input-sm"
              style={{ width: 100 }}
              placeholder="开始 00:00:00"
              value={task.params.startTime}
              onChange={e => update({
                params: { ...task.params, startTime: e.target.value }
              })}
            />
            <span className="param-unit">至</span>
            <input
              type="text"
              className="input-sm"
              style={{ width: 100 }}
              placeholder="结束 00:00:00"
              value={task.params.endTime}
              onChange={e => update({
                params: { ...task.params, endTime: e.target.value }
              })}
            />
          </div>
        )}
        {task.mode === 'total-frames' && (
          <div className="param-row">
            <span className="param-label">目标帧数</span>
            <input
              type="number"
              className="input-sm"
              value={task.params.totalFrames || 20}
              min={1}
              onChange={e => update({
                params: { ...task.params, totalFrames: parseInt(e.target.value) || 1 }
              })}
            />
            <span className="param-unit">帧</span>
          </div>
        )}
        {task.mode === 'scene-change' && (
          <div className="param-row">
            <span className="param-label">灵敏度</span>
            <input
              type="range"
              min={0.1}
              max={0.9}
              step={0.05}
              value={task.params.sceneThreshold}
              onChange={e => update({
                params: { ...task.params, sceneThreshold: parseFloat(e.target.value) }
              })}
            />
            <span className="param-value">{task.params.sceneThreshold.toFixed(2)}</span>
          </div>
        )}
        {task.mode === 'timestamps' && (
          <div className="param-timestamps">
            {(task.params.timestamps || []).map((ts, i) => (
              <div key={i} className="param-row">
                <input
                  type="text"
                  className="input-sm"
                  style={{ width: 120 }}
                  placeholder="00:00:00"
                  value={ts}
                  onChange={e => {
                    const timestamps = [...(task.params.timestamps || [])]
                    timestamps[i] = e.target.value
                    update({ params: { ...task.params, timestamps } })
                  }}
                />
                <button
                  className="btn-icon"
                  onClick={() => {
                    const timestamps = (task.params.timestamps || []).filter((_, j) => j !== i)
                    update({ params: { ...task.params, timestamps } })
                  }}
                >✕</button>
              </div>
            ))}
            <button
              className="btn btn-secondary"
              onClick={() => {
                const timestamps = [...(task.params.timestamps || []), '']
                update({ params: { ...task.params, timestamps } })
              }}
            >+ 添加时间点</button>
          </div>
        )}
      </div>
    </div>
  )
}
