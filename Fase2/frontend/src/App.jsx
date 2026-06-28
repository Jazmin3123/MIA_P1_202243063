import { useCallback, useEffect, useMemo, useState } from 'react'
import './App.css'

const API_URL = 'http://localhost:8080'

const initialCommand = 'mkdisk -size=10 -unit=M -path=/tmp/disco.mia'

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
  }, [fetchMounts])

  useEffect(() => {
    if (selectedMountId) {
      loadTree(selectedMountId, '/')
    }
  }, [loadTree, selectedMountId])

  async function executeCommand() {
    setRunningCommand(true)
    setCommandResult('')
    setCommandError('')

    try {
      const response = await fetch(`${API_URL}/execute`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command }),
      })
      const data = await parseJSONResponse(response)
      setCommandResult(data.output || 'comando ejecutado')
      await fetchMounts()
    } catch (error) {
      setCommandError(error.message)
    } finally {
      setRunningCommand(false)
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
