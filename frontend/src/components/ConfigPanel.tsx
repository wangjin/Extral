import { Task } from '../hooks/useTaskQueue'
import ModeSelector from './ModeSelector'
import OutputFormat from './OutputFormat'
import SizeConfig from './SizeConfig'
import NamingSelector from './NamingSelector'

interface Props {
  task: Task
  onUpdateTask: (task: Task) => void
  onSelectOutputDir: () => Promise<string>
}

export default function ConfigPanel({ task, onUpdateTask, onSelectOutputDir }: Props) {
  const isRunning = task.status === 'running'

  return (
    <div className="config-panel">
      <ModeSelector task={task} onChange={onUpdateTask} />
      <OutputFormat task={task} onChange={onUpdateTask} />
      <SizeConfig task={task} onChange={onUpdateTask} />
      <div className="config-section">
        <label className="config-label">保存路径</label>
        <div className="param-row">
          <input
            type="text"
            className="input-path"
            value={task.outputDir}
            onChange={e => onUpdateTask({ ...task, outputDir: e.target.value })}
            readOnly={isRunning}
          />
          <button
            className="btn btn-secondary"
            disabled={isRunning}
            onClick={async () => {
              const dir = await onSelectOutputDir()
              if (dir) {
                onUpdateTask({ ...task, outputDir: dir })
              }
            }}
          >浏览</button>
        </div>
      </div>
      <NamingSelector task={task} onChange={onUpdateTask} />
    </div>
  )
}
