import { useCallback, useEffect, useMemo, useState } from 'react'
import './App.css'

const API_URL = 'http://localhost:8080'
const REPORTS_DIR = '/home/jazmin/MIA_P1_202243063/Fase2/frontend/public/reportes'

const initialCommand = 'mkdisk -size=10 -unit=M -path=/tmp/disco.mia'
const reportOptions = [
  { label: 'Disk', name: 'disk', extension: 'svg', preview: 'image' },
  { label: 'Tree', name: 'tree', extension: 'svg', preview: 'image' },
  { label: 'Inodos', name: 'inode', extension: 'svg', preview: 'image' },
  { label: 'Bloques', name: 'block', extension: 'svg', preview: 'image' },
  { label: 'BM Inodos', name: 'bm_inode', extension: 'txt', preview: 'text' },
  { label: 'BM Bloques', name: 'bm_block', extension: 'txt', preview: 'text' },
]

function commandValue(value) {
  const text = String(value ?? '').trim()
  if (text.includes(' ')) {
    return `"${text.replaceAll('"', '')}"`
  }

  return text
}

function normalizePath(path) {
  if (!path || path === '/') {
    return '/'
  }

  return path.startsWith('/') ? path : `/${path}`
}

function joinPath(base, name) {
  if (base === '/') {
    return `/${name}`
  }

  return `${base}/${name}`
}

function parentPath(path) {
  if (!path || path === '/') {
    return '/'
  }

  const parts = path.split('/').filter(Boolean)
  parts.pop()
  return parts.length === 0 ? '/' : `/${parts.join('/')}`
}

async function parseJSONResponse(response) {
  const data = await response.json()
  if (!response.ok || data.ok === false) {
    throw new Error(data.error || 'Error al consultar el backend')
  }
  return data
}

async function executeBackendCommand(command) {
  const response = await fetch(`${API_URL}/execute`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ command }),
  })

  return parseJSONResponse(response)
}

function App() {
  const [command, setCommand] = useState(initialCommand)
  const [commandResult, setCommandResult] = useState('')
  const [commandError, setCommandError] = useState('')
  const [runningCommand, setRunningCommand] = useState(false)

  const [mounts, setMounts] = useState([])
  const [mountsError, setMountsError] = useState('')
  const [loadingMounts, setLoadingMounts] = useState(false)
  const [selectedMountId, setSelectedMountId] = useState('')

  const [currentPath, setCurrentPath] = useState('/')
  const [items, setItems] = useState([])
  const [treeError, setTreeError] = useState('')
  const [loadingTree, setLoadingTree] = useState(false)

  const [selectedFilePath, setSelectedFilePath] = useState('')
  const [fileContent, setFileContent] = useState('')
  const [fileError, setFileError] = useState('')
  const [loadingFile, setLoadingFile] = useState(false)

  const [session, setSession] = useState({ logged: false })
  const [sessionError, setSessionError] = useState('')
  const [loadingSession, setLoadingSession] = useState(false)
  const [loginPartitionId, setLoginPartitionId] = useState('')
  const [loginUser, setLoginUser] = useState('root')
  const [loginPassword, setLoginPassword] = useState('123')
  const [loginMessage, setLoginMessage] = useState('')
  const [loginError, setLoginError] = useState('')
  const [submittingLogin, setSubmittingLogin] = useState(false)

  const [reportMountId, setReportMountId] = useState('')
  const [reportName, setReportName] = useState('disk')
  const [generatingReport, setGeneratingReport] = useState(false)
  const [reportMessage, setReportMessage] = useState('')
  const [reportError, setReportError] = useState('')
  const [reportPreview, setReportPreview] = useState(null)

  const selectedMount = useMemo(
    () => mounts.find((mount) => mount.id === selectedMountId),
    [mounts, selectedMountId],
  )

  const fetchMounts = useCallback(async () => {
    setLoadingMounts(true)
    setMountsError('')

    try {
      const response = await fetch(`${API_URL}/mounts`)
      const data = await parseJSONResponse(response)
      const nextMounts = Array.isArray(data.mounts) ? data.mounts : []
      setMounts(nextMounts)

      if (!selectedMountId && nextMounts.length > 0) {
        setSelectedMountId(nextMounts[0].id)
      }
    } catch (error) {
      setMountsError(error.message)
    } finally {
      setLoadingMounts(false)
    }
  }, [selectedMountId])

  const fetchSession = useCallback(async () => {
    setLoadingSession(true)
    setSessionError('')

    try {
      const response = await fetch(`${API_URL}/session`)
      const data = await parseJSONResponse(response)
      setSession(data.session || { logged: false })
    } catch (error) {
      setSessionError(error.message)
      setSession({ logged: false })
    } finally {
      setLoadingSession(false)
    }
  }, [])

  const loadTree = useCallback(async (mountId, path = '/') => {
    if (!mountId) {
      setItems([])
      setTreeError('Selecciona una particion montada')
      return
    }

    const nextPath = normalizePath(path)
    setLoadingTree(true)
    setTreeError('')
    setFileError('')
    setSelectedFilePath('')
    setFileContent('')

    try {
      const params = new URLSearchParams({ id: mountId, path: nextPath })
      const response = await fetch(`${API_URL}/fs/tree?${params.toString()}`)
      const data = await parseJSONResponse(response)
      setCurrentPath(data.path || nextPath)
      setItems(Array.isArray(data.items) ? data.items : [])
    } catch (error) {
      setTreeError(error.message)
      setItems([])
    } finally {
      setLoadingTree(false)
    }
  }, [])

  const readFile = useCallback(async (mountId, path) => {
    if (!mountId) {
      setFileError('Selecciona una particion montada')
      return
    }

    setLoadingFile(true)
    setFileError('')
    setSelectedFilePath(path)
    setFileContent('')

    try {
      const params = new URLSearchParams({ id: mountId, path })
      const response = await fetch(`${API_URL}/fs/file?${params.toString()}`)
      const data = await parseJSONResponse(response)
      setFileContent(data.content || '')
    } catch (error) {
      setFileError(error.message)
    } finally {
      setLoadingFile(false)
    }
  }, [])

  useEffect(() => {
    fetchMounts()
    fetchSession()
  }, [fetchMounts, fetchSession])

  useEffect(() => {
    if (selectedMountId) {
      loadTree(selectedMountId, '/')
    }
  }, [loadTree, selectedMountId])

  useEffect(() => {
    if (!loginPartitionId && selectedMountId) {
      setLoginPartitionId(selectedMountId)
    }
  }, [loginPartitionId, selectedMountId])

  useEffect(() => {
    if (!reportMountId && selectedMountId) {
      setReportMountId(selectedMountId)
    }
  }, [reportMountId, selectedMountId])

  async function executeCommand() {
    setRunningCommand(true)
    setCommandResult('')
    setCommandError('')

    try {
      const data = await executeBackendCommand(command)
      setCommandResult(data.output || 'comando ejecutado')
      await fetchMounts()
      await fetchSession()
    } catch (error) {
      setCommandError(error.message)
    } finally {
      setRunningCommand(false)
    }
  }

  async function submitLogin(event) {
    event.preventDefault()
    setSubmittingLogin(true)
    setLoginMessage('')
    setLoginError('')

    const loginCommand = `login -user=${commandValue(loginUser)} -pass=${commandValue(
      loginPassword,
    )} -id=${commandValue(loginPartitionId)}`

    try {
      const data = await executeBackendCommand(loginCommand)
      setLoginMessage(data.output || 'sesion iniciada')
      await fetchSession()
    } catch (error) {
      setLoginError(error.message)
      await fetchSession()
    } finally {
      setSubmittingLogin(false)
    }
  }

  async function submitLogout() {
    setSubmittingLogin(true)
    setLoginMessage('')
    setLoginError('')

    try {
      const data = await executeBackendCommand('logout')
      setLoginMessage(data.output || 'sesion cerrada')
      await fetchSession()
    } catch (error) {
      setLoginError(error.message)
      await fetchSession()
    } finally {
      setSubmittingLogin(false)
    }
  }

  async function generateReport(event) {
    event.preventDefault()
    setGeneratingReport(true)
    setReportMessage('')
    setReportError('')
    setReportPreview(null)

    const report = reportOptions.find((option) => option.name === reportName) || reportOptions[0]
    const stamp = Date.now()
    const fileName = `${report.name}_${reportMountId || 'sin_id'}_${stamp}.${report.extension}`
    const outputPath = `${REPORTS_DIR}/${fileName}`
    const publicUrl = `/reportes/${fileName}?t=${stamp}`
    const reportCommand = `rep -id=${commandValue(reportMountId)} -name=${report.name} -path=${outputPath}`

    try {
      await executeBackendCommand(reportCommand)

      if (report.preview === 'text') {
        const response = await fetch(publicUrl)
        if (!response.ok) {
          throw new Error('No se pudo cargar el reporte generado')
        }
        const text = await response.text()
        setReportPreview({ type: 'text', content: text, path: outputPath })
      } else {
        setReportPreview({ type: 'image', url: publicUrl, path: outputPath })
      }

      setReportMessage(`Reporte generado: ${outputPath}`)
    } catch (error) {
      setReportError(error.message)
    } finally {
      setGeneratingReport(false)
    }
  }

  function handleItemClick(item) {
    const nextPath = joinPath(currentPath, item.name)
    if (item.type === 'folder') {
      loadTree(selectedMountId, nextPath)
      return
    }

    readFile(selectedMountId, nextPath)
  }

  return (
    <main className="app-shell">
      <header className="app-header">
        <div>
          <p className="eyebrow">MIA Proyecto Fase 2</p>
          <h1>Consola y Explorador EXT2</h1>
        </div>
        <button className="secondary-button" type="button" onClick={fetchMounts}>
          Actualizar montajes
        </button>
      </header>

      <section className="workspace-grid">
        <section className="panel console-panel">
          <div className="panel-heading">
            <div>
              <h2>Consola</h2>
              <p>Ejecuta comandos del backend sin salir del navegador.</p>
            </div>
            <span className="status-pill">POST /execute</span>
          </div>

          <textarea
            value={command}
            onChange={(event) => setCommand(event.target.value)}
            spellCheck="false"
            className="command-input"
          />

          <div className="actions-row">
            <button type="button" onClick={executeCommand} disabled={runningCommand}>
              {runningCommand ? 'Ejecutando...' : 'Ejecutar'}
            </button>
          </div>

          {(commandResult || commandError) && (
            <pre className={commandError ? 'result-box error-box' : 'result-box'}>
              {commandError || commandResult}
            </pre>
          )}
        </section>

        <section className="panel mounts-panel">
          <div className="panel-heading">
            <div>
              <h2>Montajes</h2>
              <p>Particiones disponibles para explorar.</p>
            </div>
            {loadingMounts && <span className="status-pill">Cargando</span>}
          </div>

          {mountsError && <div className="inline-error">{mountsError}</div>}

          <div className="mount-list">
            {mounts.length === 0 && !mountsError && (
              <div className="empty-state">No hay particiones montadas.</div>
            )}

            {mounts.map((mount) => (
              <button
                type="button"
                key={mount.id}
                className={mount.id === selectedMountId ? 'mount-item active' : 'mount-item'}
                onClick={() => setSelectedMountId(mount.id)}
              >
                <span className="mount-id">{mount.id}</span>
                <span className="mount-name">{mount.name}</span>
                <span className="mount-path">{mount.path}</span>
                <span className="mount-size">{mount.size || 0} bytes</span>
              </button>
            ))}
          </div>
        </section>
      </section>

      <section className="panel login-panel">
        <div className="panel-heading">
          <div>
            <h2>Login</h2>
            <p>Autentica una sesion usando el comando existente del backend.</p>
          </div>
          {loadingSession && <span className="status-pill">Consultando</span>}
        </div>

        {sessionError && <div className="inline-error">{sessionError}</div>}

        {session.logged ? (
          <div className="session-card">
            <div>
              <span className="session-label">Usuario conectado</span>
              <strong>
                {session.user} | Grupo: {session.group} | Particion: {session.partitionID}
              </strong>
            </div>
            <button type="button" onClick={submitLogout} disabled={submittingLogin}>
              {submittingLogin ? 'Cerrando...' : 'Logout'}
            </button>
          </div>
        ) : (
          <form className="login-form" onSubmit={submitLogin}>
            <label>
              <span>ID particion</span>
              <input
                value={loginPartitionId}
                onChange={(event) => setLoginPartitionId(event.target.value)}
                placeholder="631A"
              />
            </label>
            <label>
              <span>Usuario</span>
              <input
                value={loginUser}
                onChange={(event) => setLoginUser(event.target.value)}
                placeholder="root"
              />
            </label>
            <label>
              <span>Contrasena</span>
              <input
                type="password"
                value={loginPassword}
                onChange={(event) => setLoginPassword(event.target.value)}
                placeholder="123"
              />
            </label>
            <button type="submit" disabled={submittingLogin}>
              {submittingLogin ? 'Ingresando...' : 'Login'}
            </button>
          </form>
        )}

        {(loginMessage || loginError) && (
          <pre className={loginError ? 'result-box error-box' : 'result-box'}>
            {loginError || loginMessage}
          </pre>
        )}
      </section>

      <section className="panel reports-panel">
        <div className="panel-heading">
          <div>
            <h2>Reportes</h2>
            <p>Genera reportes con el comando rep y visualizalos en la GUI.</p>
          </div>
          <span className="status-pill">POST /execute</span>
        </div>

        <form className="reports-form" onSubmit={generateReport}>
          <label>
            <span>Reporte</span>
            <select value={reportName} onChange={(event) => setReportName(event.target.value)}>
              {reportOptions.map((option) => (
                <option key={option.name} value={option.name}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>

          <label>
            <span>ID particion</span>
            <select value={reportMountId} onChange={(event) => setReportMountId(event.target.value)}>
              <option value="">Selecciona un montaje</option>
              {mounts.map((mount) => (
                <option key={mount.id} value={mount.id}>
                  {mount.id} - {mount.name}
                </option>
              ))}
            </select>
          </label>

          <button type="submit" disabled={generatingReport || !reportMountId}>
            {generatingReport ? 'Generando...' : 'Generar'}
          </button>
        </form>

        {(reportMessage || reportError) && (
          <pre className={reportError ? 'result-box error-box' : 'result-box'}>
            {reportError || reportMessage}
          </pre>
        )}

        {reportPreview && (
          <div className="report-preview">
            <div className="preview-heading">
              <h3>Vista previa</h3>
              <span>{reportPreview.path}</span>
            </div>

            {reportPreview.type === 'image' ? (
              <div className="report-image-frame">
                <img src={reportPreview.url} alt="Reporte generado" />
              </div>
            ) : (
              <pre className="content-box report-text">{reportPreview.content}</pre>
            )}
          </div>
        )}
      </section>

      <section className="panel explorer-panel">
        <div className="panel-heading explorer-heading">
          <div>
            <h2>Explorador EXT2</h2>
            <p>{selectedMount ? `${selectedMount.id} - ${selectedMount.name}` : 'Selecciona un montaje'}</p>
          </div>
          <div className="path-controls">
            <span className="path-chip">{currentPath}</span>
            <button
              className="secondary-button"
              type="button"
              disabled={!selectedMountId || currentPath === '/'}
              onClick={() => loadTree(selectedMountId, parentPath(currentPath))}
            >
              Subir
            </button>
          </div>
        </div>

        {treeError && <div className="inline-error">{treeError}</div>}

        <div className="explorer-grid">
          <div className="file-list">
            {loadingTree && <div className="empty-state">Cargando carpeta...</div>}

            {!loadingTree && items.length === 0 && !treeError && (
              <div className="empty-state">La carpeta esta vacia.</div>
            )}

            {items.map((item) => (
              <button
                type="button"
                key={`${item.inode}-${item.name}`}
                className="file-row"
                onClick={() => handleItemClick(item)}
              >
                <span className={item.type === 'folder' ? 'file-icon folder-icon' : 'file-icon'}>
                  {item.type === 'folder' ? 'DIR' : 'TXT'}
                </span>
                <span className="file-name">{item.name}</span>
                <span className="file-meta">inodo {item.inode}</span>
              </button>
            ))}
          </div>

          <aside className="file-preview">
            <div className="preview-heading">
              <h3>Contenido</h3>
              <span>{loadingFile ? 'Leyendo...' : selectedFilePath || 'Sin archivo'}</span>
            </div>

            {fileError && <div className="inline-error">{fileError}</div>}

            <pre className="content-box">
              {fileContent || 'Selecciona un archivo para ver su contenido.'}
            </pre>
          </aside>
        </div>
      </section>
    </main>
  )
}

export default App
