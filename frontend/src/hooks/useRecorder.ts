import { useState, useEffect, useCallback, useRef } from 'react'
import { Events } from '@wailsio/runtime'
import * as RecorderService from '../../bindings/video-extractor/internal/recorder/service'
import { RecordingConfig } from '../../bindings/video-extractor/internal/recorder/models'

export interface Camera {
  id: string
  name: string
}

export interface AudioDevice {
  id: string
  name: string
}

const defaultConfig: RecordingConfig = {
  cameraId: '',
  audioDevice: '',
  resolution: '1280x720',
  fps: 30,
  codec: 'libx264',
  rateControl: 'crf',
  quality: 'medium',
  crf: 23,
  bitrate: 5000,
  format: 'mp4',
  maxDuration: 0,
  outputDir: '',
  fileName: '',
}

export function useRecorder() {
  const [cameras, setCameras] = useState<Camera[]>([])
  const [audioDevices, setAudioDevices] = useState<AudioDevice[]>([])
  const [selectedCamera, setSelectedCamera] = useState('')
  const [selectedAudio, setSelectedAudio] = useState('')
  const [config, setConfig] = useState<RecordingConfig>(defaultConfig)
  const [previewFrame, setPreviewFrame] = useState('')
  const [isPreviewing, setIsPreviewing] = useState(false)
  const [isRecording, setIsRecording] = useState(false)
  const [isPaused, setIsPaused] = useState(false)
  const [duration, setDuration] = useState(0)
  const [fileSize, setFileSize] = useState(0)
  const [fps, setFps] = useState(0)
  const [dropped, setDropped] = useState(0)
  const [logs, setLogs] = useState<string[]>([])
  const [error, setError] = useState<string | null>(null)

  const mountedRef = useRef(true)

  useEffect(() => {
    mountedRef.current = true
    return () => { mountedRef.current = false }
  }, [])

  // Load devices and listen to events on mount
  useEffect(() => {
    loadDevices()

    const off1 = Events.On('recorder:preview-frame', (ev: any) => {
      setPreviewFrame(ev.data || ev)
    })

    const off2 = Events.On('recorder:status', (ev: any) => {
      const status = ev.data || ev
      if (status.state === 'recording') {
        setIsRecording(true)
        setIsPaused(false)
      } else if (status.state === 'paused') {
        setIsPaused(true)
      } else {
        setIsRecording(false)
        setIsPaused(false)
      }
      if (status.duration !== undefined) setDuration(status.duration)
      if (status.fileSize !== undefined) setFileSize(status.fileSize)
      if (status.fps !== undefined) setFps(status.fps)
      if (status.dropped !== undefined) setDropped(status.dropped)
    })

    const off3 = Events.On('recorder:log', (ev: any) => {
      const line = ev.data || ev
      setLogs(prev => [...prev, line])
    })

    const off4 = Events.On('recorder:error', (ev: any) => {
      const data = ev.data || ev
      setError(data.error || String(data))
    })

    return () => {
      off1()
      off2()
      off3()
      off4()
    }
  }, [])

  const loadDevices = useCallback(async () => {
    try {
      const cams = await RecorderService.ListCameras()
      setCameras(cams || [])
      if (cams && cams.length > 0) {
        setSelectedCamera(cams[0].id)
      }
    } catch (e) {
      console.error('ListCameras failed:', e)
    }
    try {
      const devs = await RecorderService.ListAudioDevices()
      setAudioDevices(devs || [])
    } catch (e) {
      console.error('ListAudioDevices failed:', e)
    }
  }, [])

  // Start preview when camera is selected and not recording
  useEffect(() => {
    if (!selectedCamera || isRecording) return

    setPreviewFrame('')
    RecorderService.StartPreview(selectedCamera).then(() => {
      setIsPreviewing(true)
    }).catch(e => {
      console.error('StartPreview failed:', e)
      setError(String(e))
    })

    return () => {
      RecorderService.StopPreview().catch(() => {})
    }
  }, [selectedCamera])

  const selectCamera = useCallback((id: string) => {
    setSelectedCamera(id)
    setConfig(prev => ({ ...prev, cameraId: id }))
  }, [])

  const selectAudio = useCallback((id: string) => {
    setSelectedAudio(id)
    setConfig(prev => ({ ...prev, audioDevice: id }))
  }, [])

  const updateConfig = useCallback((updates: Partial<RecordingConfig>) => {
    setConfig(prev => ({ ...prev, ...updates }))
  }, [])

  const startRecording = useCallback(async () => {
    setLogs([])
    setError(null)
    try {
      const fullConfig: RecordingConfig = {
        ...config,
        cameraId: selectedCamera,
        audioDevice: selectedAudio,
      }
      if (!fullConfig.outputDir) {
        const dir = await RecorderService.SelectOutputDir()
        if (dir) {
          fullConfig.outputDir = dir
          setConfig(prev => ({ ...prev, outputDir: dir }))
        }
      }
      await RecorderService.StartRecording(fullConfig)
      setIsRecording(true)
      setIsPreviewing(false)
    } catch (e) {
      setError(String(e))
    }
  }, [config, selectedCamera, selectedAudio])

  const stopRecording = useCallback(async () => {
    try {
      await RecorderService.StopRecording()
      setIsRecording(false)
      setIsPaused(false)
      setDuration(0)
      setFileSize(0)
      setFps(0)
      // Restart preview
      if (selectedCamera) {
        RecorderService.StartPreview(selectedCamera).then(() => {
          setIsPreviewing(true)
        }).catch(() => {})
      }
    } catch (e) {
      setError(String(e))
    }
  }, [selectedCamera])

  const pauseRecording = useCallback(async () => {
    try {
      await RecorderService.PauseRecording()
      setIsPaused(true)
    } catch (e) {
      setError(String(e))
    }
  }, [])

  const resumeRecording = useCallback(async () => {
    try {
      await RecorderService.ResumeRecording()
      setIsPaused(false)
    } catch (e) {
      setError(String(e))
    }
  }, [])

  const selectOutputDir = useCallback(async () => {
    const dir = await RecorderService.SelectOutputDir()
    if (dir) {
      setConfig(prev => ({ ...prev, outputDir: dir }))
    }
  }, [])

  return {
    cameras,
    audioDevices,
    selectedCamera,
    selectedAudio,
    config,
    previewFrame,
    isPreviewing,
    isRecording,
    isPaused,
    duration,
    fileSize,
    fps,
    dropped,
    logs,
    error,
    selectCamera,
    selectAudio,
    updateConfig,
    startRecording,
    stopRecording,
    pauseRecording,
    resumeRecording,
    selectOutputDir,
    loadDevices,
  }
}
