import { useState, useEffect, useCallback } from 'react'
import { Events } from '@wailsio/runtime'
import * as Service from '../../bindings/video-extractor/internal/app/service'
import { Task } from '../../bindings/video-extractor/internal/model/models'

export type { Task }

export function useTaskQueue() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null)

  useEffect(() => {
    Service.GetTasks().then(result => {
      setTasks((result || []).filter((t): t is Task => t !== null))
    }).catch(() => {})

    Events.On('task:changed', (ev: any) => {
      const task: Task = ev.data || ev
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
    const files = await Service.SelectFiles()
    if (files && files.length > 0) {
      const newTasks = await Service.AddTasks(files)
      if (newTasks && newTasks.length > 0) {
        const first = newTasks[0]
        if (first) setSelectedTaskId(first.id)
      }
    }
  }, [])

  const addFolder = useCallback(async () => {
    const folder = await Service.SelectFolder()
    if (folder) {
      const files = await Service.ScanVideoFiles(folder)
      if (files && files.length > 0) {
        const newTasks = await Service.AddTasks(files)
        if (newTasks && newTasks.length > 0) {
          const first = newTasks[0]
          if (first) setSelectedTaskId(first.id)
        }
      }
    }
  }, [])

  const updateTask = useCallback(async (task: Task) => {
    await Service.UpdateTask(task)
  }, [])

  const removeTask = useCallback(async (id: string) => {
    await Service.RemoveTask(id)
    setSelectedTaskId(prev => prev === id ? null : prev)
  }, [])

  const clearTasks = useCallback(async () => {
    await Service.ClearTasks()
    setSelectedTaskId(null)
  }, [])

  const startAll = useCallback(async () => {
    await Service.StartAll()
  }, [])

  const cancelTask = useCallback(async (id: string) => {
    await Service.CancelTask(id)
  }, [])

  const selectOutputDir = useCallback(async (): Promise<string> => {
    return await Service.SelectOutputDir()
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
