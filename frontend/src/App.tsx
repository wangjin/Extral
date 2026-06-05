import './App.css'
import TaskList from './components/TaskList'
import ConfigPanel from './components/ConfigPanel'
import ActionBar from './components/ActionBar'
import UpdateToast from './components/UpdateToast'
import UpdateModal from './components/UpdateModal'
import RecorderPanel from './components/RecorderPanel'
import { useTaskQueue } from './hooks/useTaskQueue'
import { useUpdate } from './hooks/useUpdate'
import { useState, useCallback, useEffect } from 'react'
import * as AppService from '../bindings/video-extractor/internal/app/service'
import * as UpdaterService from '../bindings/video-extractor/internal/updater/service'

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

  const { status, info, progress, error, ignoreUpdate } = useUpdate()
  const [showUpdateModal, setShowUpdateModal] = useState(false)
  const [showLogs, setShowLogs] = useState(false)
  const [version, setVersion] = useState('')
  const [activeTab, setActiveTab] = useState<'extract' | 'record'>('extract')

  useEffect(() => {
    AppService.GetVersion().then(v => setVersion(v)).catch(() => {})
  }, [])

  const selectedTask = tasks.find(t => t.id === selectedTaskId)

  const handleViewDetails = useCallback(() => {
    setShowUpdateModal(true)
  }, [])

  const handleStartDownload = useCallback(async () => {
    try { await UpdaterService.StartDownload() } catch (e) { console.error('StartDownload failed:', e) }
  }, [])

  const handleCancelDownload = useCallback(async () => {
    await UpdaterService.CancelDownload()
    setShowUpdateModal(false)
  }, [])

  const handleInstall = useCallback(async () => {
    await UpdaterService.InstallAndRestart()
  }, [])

  const handleManualDownload = useCallback(async () => {
    setShowUpdateModal(false)
    await UpdaterService.OpenReleasePage()
  }, [])

  const handleDismissToast = useCallback(() => {
    ignoreUpdate()
  }, [ignoreUpdate])

  const handleDismissModal = useCallback(() => {
    if (status !== 'downloading') {
      setShowUpdateModal(false)
    }
  }, [status])

  return (
    <div className="app">
      <header className="toolbar">
        <div className="toolbar-brand">
          <img className="toolbar-logo" src="/logo.png" alt="Extral" />
          <div className="toolbar-title">Extral</div>
          {version && <span className="toolbar-version">{version}</span>}
        </div>
        <div className="toolbar-tabs">
          <button
            className={`toolbar-tab ${activeTab === 'extract' ? 'active' : ''}`}
            onClick={() => setActiveTab('extract')}
          >抽帧</button>
          <button
            className={`toolbar-tab ${activeTab === 'record' ? 'active' : ''}`}
            onClick={() => setActiveTab('record')}
          >录制</button>
        </div>
        {activeTab === 'extract' ? (
          <div className="toolbar-actions">
            <button className="btn btn-primary" onClick={addFiles}>选择视频</button>
            <button className="btn btn-secondary" onClick={addFolder}>选择文件夹</button>
          </div>
        ) : (
          <div className="toolbar-actions" />
        )}
      </header>

      {activeTab === 'extract' ? (
        <>
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

          {selectedTask && selectedTask.status !== 'pending' && (
            <div className="log-panel">
              <div className="log-header" onClick={() => setShowLogs(!showLogs)}>
                <span className="log-header-title">执行日志 ({(selectedTask.logs || []).length})</span>
                <span className="log-toggle">{showLogs ? '收起' : '展开'}</span>
              </div>
              {showLogs && (
                <pre className="log-content">
                  {(selectedTask.logs || []).map((line, i) => (
                    <div key={i} className={`log-line ${line.startsWith('[命令') || line.startsWith('[错误') || line.startsWith('[完成') ? 'log-line-highlight' : ''}`}>
                      {line}
                    </div>
                  ))}
                </pre>
              )}
            </div>
          )}

          <ActionBar
            tasks={tasks}
            onStartAll={startAll}
            onClear={clearTasks}
          />
        </>
      ) : (
        <RecorderPanel />
      )}

      {status === 'available' && !showUpdateModal && info && (
        <UpdateToast info={info} onViewDetails={handleViewDetails} onDismiss={handleDismissToast} />
      )}
      <UpdateModal
        visible={showUpdateModal}
        status={status}
        info={info}
        progress={progress}
        error={error}
        onStartDownload={handleStartDownload}
        onCancel={handleCancelDownload}
        onInstall={handleInstall}
        onManualDownload={handleManualDownload}
        onClose={handleDismissModal}
      />
    </div>
  )
}

export default App
