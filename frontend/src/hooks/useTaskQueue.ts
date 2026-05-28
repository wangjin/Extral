import { useState, useEffect, useCallback } from 'react'

export interface ModeParams {
  frameInterval: number
  secondInterval: number
  startTime: string
  endTime: string
  totalFrames: number
  sceneThreshold: number
  timestamps: string[]
}

export interface Task {
  id: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
  videoPath: string
  mode: string
  params: ModeParams
  format: string
  quality: number
  scaleMode: string
  scale: number
  width: number
  height: number
  keepRatio: boolean
  outputDir: string
  naming: string
  progress: number
  current: number
  total: number
  elapsedTime: number
  error: string
  createdAt: string
}

declare function AddTasks(paths: string[]): Promise<Task[]>
declare function GetTasks(): Promise<Task[]>
declare function UpdateTask(task: Task): Promise<void>
declare function RemoveTask(id: string): Promise<void>
declare function ClearTasks(): Promise<void>
declare function SelectFiles(): Promise<string[]>
declare function SelectFolder(): Promise<string>
declare function SelectOutputDir(): Promise<string>
declare function StartAll(): Promise<void>
declare function CancelTask(id: string): Promise<void>
declare function ScanVideoFiles(dir: string): Promise<string[]>
declare function EventsOn(event: string, callback: (...args: any[]) => void): void

export function useTaskQueue() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null)

  useEffect(() => {
    GetTasks().then(setTasks).catch(() => {})

    EventsOn('task:changed', (task: Task) => {
      setTasks(prev => {
        const index = prev.findIndex(t => t.id === task.id)
        if (index >= 0) {
          const next = [...prev]
          next[index] = task
          return next
        }
        return [...prev, task]
      })
    })
  }, [])

  const selectTask = useCallback((id: string | null) => {
    setSelectedTaskId(id)
  }, [])

  const addFiles = useCallback(async () => {
    const files = await SelectFiles()
    if (files && files.length > 0) {
      const newTasks = await AddTasks(files)
      if (newTasks.length > 0) {
        setSelectedTaskId(newTasks[0].id)
      }
    }
  }, [])

  const addFolder = useCallback(async () => {
    const folder = await SelectFolder()
    if (folder) {
      const files = await ScanVideoFiles(folder)
      if (files.length > 0) {
        const newTasks = await AddTasks(files)
        if (newTasks.length > 0) {
          setSelectedTaskId(newTasks[0].id)
        }
      }
    }
  }, [])

  const updateTask = useCallback(async (task: Task) => {
    await UpdateTask(task)
  }, [])

  const removeTask = useCallback(async (id: string) => {
    await RemoveTask(id)
    setSelectedTaskId(prev => prev === id ? null : prev)
  }, [])

  const clearTasks = useCallback(async () => {
    await ClearTasks()
    setSelectedTaskId(null)
  }, [])

  const startAll = useCallback(async () => {
    await StartAll()
  }, [])

  const cancelTask = useCallback(async (id: string) => {
    await CancelTask(id)
  }, [])

  const selectOutputDir = useCallback(async (): Promise<string> => {
    return await SelectOutputDir()
  }, [])

  return {
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
  }
}
