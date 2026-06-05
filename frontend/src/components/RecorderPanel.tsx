import { useRecorder } from '../hooks/useRecorder'

export default function RecorderPanel() {
  const {
    cameras, audioDevices,
    selectedCamera, selectedAudio,
    config, previewFrame,
    isRecording, isPaused,
    duration, fileSize, fps,
    logs, error,
    selectCamera, selectAudio,
    updateConfig,
    startRecording, stopRecording,
    pauseRecording, resumeRecording,
    selectOutputDir,
  } = useRecorder()

  const formatDuration = (sec: number) => {
    const h = Math.floor(sec / 3600)
    const m = Math.floor((sec % 3600) / 60)
    const s = Math.floor(sec % 60)
    return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  }

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / 1024 / 1024).toFixed(1)} MB`
  }

  return (
    <div className="recorder-panel">
      {/* Left: Preview + Controls */}
      <div className="recorder-left">
        <div className="recorder-preview">
          {isRecording ? (
            <div className="recorder-recording-overlay">
              <div className="recording-indicator">
                <span className="recording-dot" />
                {isPaused ? '已暂停' : '录制中'}
              </div>
              <div className="recording-time">{formatDuration(duration)}</div>
            </div>
          ) : previewFrame ? (
            <img src={`data:image/jpeg;base64,${previewFrame}`} alt="预览" />
          ) : (
            <div className="recorder-preview-empty">
              <span className="preview-icon">📷</span>
              <span>{cameras.length === 0 ? '未检测到摄像头设备' : '选择摄像头开始预览'}</span>
            </div>
          )}
        </div>

        <div className="recorder-controls">
          {!isRecording ? (
            <button
              className="btn btn-primary recorder-btn-record"
              disabled={!selectedCamera}
              onClick={startRecording}
            >
              ● 开始录制
            </button>
          ) : (
            <>
              {isPaused ? (
                <button className="btn btn-secondary" onClick={resumeRecording}>
                  ▶ 继续
                </button>
              ) : (
                <button className="btn btn-secondary" onClick={pauseRecording}>
                  ⏸ 暂停
                </button>
              )}
              <button className="btn btn-danger" onClick={stopRecording}>
                ⏹ 停止
              </button>
            </>
          )}
        </div>

        <div className="recorder-status">
          <span>⏱ {formatDuration(duration)}</span>
          <span>📁 {formatSize(fileSize)}</span>
          <span>🎬 {fps.toFixed(0)} fps</span>
        </div>
      </div>

      {/* Right: Config Panel */}
      <div className="recorder-right">
        <div className="config-section">
          <label className="config-label">摄像头</label>
          <select
            className="recorder-select"
            value={selectedCamera}
            onChange={e => selectCamera(e.target.value)}
            disabled={isRecording}
          >
            <option value="">选择摄像头</option>
            {cameras.map(c => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
        </div>

        <div className="config-section">
          <label className="config-label">麦克风</label>
          <select
            className="recorder-select"
            value={selectedAudio}
            onChange={e => selectAudio(e.target.value)}
            disabled={isRecording}
          >
            <option value="">不录制音频</option>
            {audioDevices.map(d => (
              <option key={d.id} value={d.id}>{d.name}</option>
            ))}
          </select>
        </div>

        <div className="config-section">
          <label className="config-label">分辨率</label>
          <div className="btn-group">
            {[
              { label: '1080p', value: '1920x1080' },
              { label: '720p', value: '1280x720' },
              { label: '480p', value: '854x480' },
            ].map(r => (
              <button
                key={r.value}
                className={`btn-option ${config.resolution === r.value ? 'active' : ''}`}
                disabled={isRecording}
                onClick={() => updateConfig({ resolution: r.value })}
              >{r.label}</button>
            ))}
          </div>
        </div>

        <div className="config-section">
          <label className="config-label">帧率</label>
          <div className="btn-group">
            {[30, 60].map(f => (
              <button
                key={f}
                className={`btn-option ${config.fps === f ? 'active' : ''}`}
                disabled={isRecording}
                onClick={() => updateConfig({ fps: f })}
              >{f} fps</button>
            ))}
          </div>
        </div>

        <div className="config-section">
          <label className="config-label">质量控制</label>
          <div className="btn-group" style={{ marginBottom: '6px' }}>
            {(['crf', 'bitrate'] as const).map(mode => (
              <button
                key={mode}
                className={`btn-option ${config.rateControl === mode ? 'active' : ''}`}
                disabled={isRecording}
                onClick={() => updateConfig({ rateControl: mode })}
              >{mode === 'crf' ? 'CRF 质量' : '码率'}</button>
            ))}
          </div>
          {config.rateControl === 'crf' ? (
            <>
              <div className="btn-group" style={{ marginBottom: '6px' }}>
                {[
                  { label: '低', value: 'low' },
                  { label: '中', value: 'medium' },
                  { label: '高', value: 'high' },
                  { label: '自定义', value: 'custom' },
                ].map(q => (
                  <button
                    key={q.value}
                    className={`btn-option ${config.quality === q.value ? 'active' : ''}`}
                    disabled={isRecording}
                    onClick={() => updateConfig({ quality: q.value })}
                  >{q.label}</button>
                ))}
              </div>
              {config.quality === 'custom' && (
                <div className="param-row">
                  <span className="param-label">CRF</span>
                  <input
                    type="range" min="0" max="51"
                    value={config.crf}
                    onChange={e => updateConfig({ crf: Number(e.target.value) })}
                  />
                  <span className="param-value">{config.crf}</span>
                </div>
              )}
            </>
          ) : (
            <div className="param-row">
              <input
                type="number" className="input-sm"
                value={config.bitrate}
                onChange={e => updateConfig({ bitrate: Number(e.target.value) })}
                disabled={isRecording}
              />
              <span className="param-unit">kbps</span>
            </div>
          )}
        </div>

        <div className="config-section">
          <label className="config-label">输出格式</label>
          <div className="btn-group">
            {['mp4', 'mkv'].map(f => (
              <button
                key={f}
                className={`btn-option ${config.format === f ? 'active' : ''}`}
                disabled={isRecording}
                onClick={() => updateConfig({ format: f })}
              >{f.toUpperCase()}</button>
            ))}
          </div>
        </div>

        <div className="config-section">
          <label className="config-label">最大时长</label>
          <div className="param-row">
            <input
              type="number" className="input-sm"
              value={config.maxDuration || ''}
              placeholder="0 = 不限"
              onChange={e => updateConfig({ maxDuration: Number(e.target.value) || 0 })}
              disabled={isRecording}
            />
            <span className="param-unit">秒 (0=不限)</span>
          </div>
        </div>

        <div className="config-section">
          <label className="config-label">保存路径</label>
          <div className="param-row">
            <input
              type="text" className="input-path"
              value={config.outputDir}
              onChange={e => updateConfig({ outputDir: e.target.value })}
              readOnly={isRecording}
              placeholder="点击浏览选择"
            />
            <button
              className="btn btn-secondary"
              disabled={isRecording}
              onClick={selectOutputDir}
            >浏览</button>
          </div>
        </div>
      </div>

      {/* Error */}
      {error && (
        <div className="recorder-error">{error}</div>
      )}

      {/* Log Panel */}
      {isRecording && logs.length > 0 && (
        <div className="log-panel" style={{ gridColumn: '1 / -1' }}>
          <div className="log-header">
            <span className="log-header-title">录制日志 ({logs.length})</span>
          </div>
          <pre className="log-content">
            {logs.slice(-50).map((line, i) => (
              <div key={i} className={`log-line ${line.startsWith('[命令') || line.startsWith('[错误') || line.startsWith('[完成') ? 'log-line-highlight' : ''}`}>
                {line}
              </div>
            ))}
          </pre>
        </div>
      )}
    </div>
  )
}
