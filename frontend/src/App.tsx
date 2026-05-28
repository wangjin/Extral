import './App.css'
import TaskList from './components/TaskList'
import ConfigPanel from './components/ConfigPanel'
import ActionBar from './components/ActionBar'
import { useTaskQueue } from './hooks/useTaskQueue'

function App() {
  const {
    tasks,
    selectedTaskId,
    selectTask,
    addFiles,
    addFolder,
    updateTask,
    removeTask,
    clearTasks,
    startAll,
    cancelTask,
    selectOutputDir,
  } = useTaskQueue()

  const selectedTask = tasks.find(t => t.id === selectedTaskId)

  return (
    <div className="app">
      <header className="toolbar">
        <div className="toolbar-title">视频抽帧工具</div>
        <div className="toolbar-actions">
          <button className="btn btn-primary" onClick={addFiles}>选择视频</button>
          <button className="btn btn-secondary" onClick={addFolder}>选择文件夹</button>
        </div>
      </header>

      <TaskList
        tasks={tasks}
        selectedTaskId={selectedTaskId}
        onSelectTask={selectTask}
        onRemoveTask={removeTask}
        onCancelTask={cancelTask}
      />

      {selectedTask && (
        <ConfigPanel
          task={selectedTask}
          onUpdateTask={updateTask}
          onSelectOutputDir={selectOutputDir}
        />
      )}

      <ActionBar
        tasks={tasks}
        onStartAll={startAll}
        onClear={clearTasks}
      />
    </div>
  )
}

export default App
