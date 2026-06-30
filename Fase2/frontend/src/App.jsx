import { useCallback, useEffect, useMemo, useState } from 'react'
import './App.css'

const API_URL = 'http://52.91.161.35:8080'
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
  if (/\s/.test(text)) {
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
  const [activeUtilityTab, setActiveUtilityTab] = useState('disks')
  const [pathHistory, setPathHistory] = useState([])
  const [runningVisualAction, setRunningVisualAction] = useState('')
  const [visualResults, setVisualResults] = useState({})
  const [visualForms, setVisualForms] = useState({
    createDisk: { size: '10', unit: 'M', fit: 'FF', path: '/tmp/disco.mia' },
    deleteDisk: { path: '/tmp/disco.mia' },
    createPartition: {
      size: '1',
      unit: 'M',
      path: '/tmp/disco.mia',
      name: 'part1',
      type: 'P',
      fit: 'FF',
    },
    mountPartition: { path: '/tmp/disco.mia', name: 'part1' },
    formatPartition: { id: '', type: 'full' },
    deletePartition: { path: '/tmp/disco.mia', name: 'part1', deleteMode: 'fast' },
    resizePartition: { path: '/tmp/disco.mia', name: 'part1', add: '1', unit: 'M' },
    createFolder: { path: '/home/docs', recursive: true },
    createFile: { path: '/home/docs/archivo.txt', content: '', size: '64' },
    renameEntry: { path: '/home/docs/archivo.txt', name: 'nuevo.txt' },
    editFile: { path: '/home/docs/archivo.txt', contentPath: '/tmp/contenido.txt' },
    removeEntry: { path: '/home/docs/archivo.txt' },
    copyEntry: { path: '/home/docs/archivo.txt', destino: '/home/copia.txt' },
    moveEntry: { path: '/home/docs/archivo.txt', destino: '/home/movido.txt' },
  })

  const selectedMount = useMemo(
    () => mounts.find((mount) => mount.id === selectedMountId),
    [mounts, selectedMountId],
  )

  const breadcrumbs = useMemo(() => {
    const mountLabel = selectedMount?.id || 'Unidad'
    const parts = currentPath.split('/').filter(Boolean)
    const crumbs = [
      { label: 'Este equipo', path: null },
      { label: mountLabel, path: '/' },
    ]

    parts.forEach((part, index) => {
      crumbs.push({ label: part, path: `/${parts.slice(0, index + 1).join('/')}` })
    })

    return crumbs
  }, [currentPath, selectedMount])

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

  useEffect(() => {
    if (selectedMountId) {
      setVisualForms((forms) => ({
        ...forms,
        formatPartition: forms.formatPartition.id
          ? forms.formatPartition
          : { ...forms.formatPartition, id: selectedMountId },
      }))
    }
  }, [selectedMountId])

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

  function updateVisualForm(formName, fieldName, value) {
    setVisualForms((forms) => ({
      ...forms,
      [formName]: {
        ...forms[formName],
        [fieldName]: value,
      },
    }))
  }

  async function executeVisualAction(event, actionKey, command, { refreshExplorer = true } = {}) {
    event.preventDefault()
    setRunningVisualAction(actionKey)
    setVisualResults((results) => ({
      ...results,
      [actionKey]: { command, output: '', error: '' },
    }))

    try {
      const data = await executeBackendCommand(command)
      setVisualResults((results) => ({
        ...results,
        [actionKey]: { command, output: data.output || 'comando ejecutado', error: '' },
      }))
      await fetchMounts()
      await fetchSession()

      if (refreshExplorer && selectedMountId) {
        await loadTree(selectedMountId, currentPath)
      }
    } catch (error) {
      setVisualResults((results) => ({
        ...results,
        [actionKey]: { command, output: '', error: error.message },
      }))
      await fetchMounts()
      await fetchSession()
    } finally {
      setRunningVisualAction('')
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

  function selectMount(mountId) {
    setPathHistory([])
    if (mountId === selectedMountId) {
      loadTree(mountId, '/')
      return
    }

    setSelectedMountId(mountId)
  }

  function navigateTo(path, { pushHistory = true } = {}) {
    if (pushHistory) {
      setPathHistory((history) => [...history, currentPath])
    }

    loadTree(selectedMountId, path)
  }

  function goBack() {
    setPathHistory((history) => {
      const previous = history[history.length - 1]
      if (previous) {
        loadTree(selectedMountId, previous)
      }

      return history.slice(0, -1)
    })
  }

  function handleItemClick(item) {
    const nextPath = joinPath(currentPath, item.name)
    if (item.type === 'folder') {
      navigateTo(nextPath)
      return
    }

    readFile(selectedMountId, nextPath)
  }

  function actionResult(actionKey) {
    const result = visualResults[actionKey]
    if (!result) {
      return null
    }

    return (
      <pre className={result.error ? 'result-box error-box' : 'result-box'}>
        {`$ ${result.command}\n${result.error || result.output}`}
      </pre>
    )
  }

  function renderTextField(formName, fieldName, label, props = {}) {
    return (
      <label>
        <span>{label}</span>
        <input
          value={visualForms[formName][fieldName]}
          onChange={(event) => updateVisualForm(formName, fieldName, event.target.value)}
          {...props}
        />
      </label>
    )
  }

  function renderSelectField(formName, fieldName, label, options) {
    return (
      <label>
        <span>{label}</span>
        <select
          value={visualForms[formName][fieldName]}
          onChange={(event) => updateVisualForm(formName, fieldName, event.target.value)}
        >
          {options.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </label>
    )
  }

  function renderActionButton(actionKey, label) {
    return (
      <button type="submit" disabled={runningVisualAction === actionKey}>
        {runningVisualAction === actionKey ? 'Ejecutando...' : label}
      </button>
    )
  }

  function renderDisksTab() {
    const createDisk = visualForms.createDisk
    const deleteDisk = visualForms.deleteDisk

    return (
      <div className="action-grid">
        <form
          className="action-card"
          onSubmit={(event) =>
            executeVisualAction(
              event,
              'createDisk',
              `mkdisk -size=${commandValue(createDisk.size)} -unit=${commandValue(
                createDisk.unit,
              )} -fit=${commandValue(createDisk.fit)} -path=${commandValue(createDisk.path)}`,
              { refreshExplorer: false },
            )
          }
        >
          <h2>Crear disco</h2>
          <div className="form-grid">
            {renderTextField('createDisk', 'size', 'Size', { type: 'number', min: '1' })}
            {renderSelectField('createDisk', 'unit', 'Unit', [
              { label: 'K', value: 'K' },
              { label: 'M', value: 'M' },
            ])}
            {renderSelectField('createDisk', 'fit', 'Fit', [
              { label: 'First Fit', value: 'FF' },
              { label: 'Best Fit', value: 'BF' },
              { label: 'Worst Fit', value: 'WF' },
            ])}
            {renderTextField('createDisk', 'path', 'Path')}
          </div>
          <div className="actions-row">{renderActionButton('createDisk', 'Crear disco')}</div>
          {actionResult('createDisk')}
        </form>

        <form
          className="action-card"
          onSubmit={(event) =>
            executeVisualAction(
              event,
              'deleteDisk',
              `rmdisk -path=${commandValue(deleteDisk.path)}`,
              { refreshExplorer: false },
            )
          }
        >
          <h2>Eliminar disco</h2>
          <div className="form-grid single">
            {renderTextField('deleteDisk', 'path', 'Path')}
          </div>
          <div className="actions-row">{renderActionButton('deleteDisk', 'Eliminar disco')}</div>
          {actionResult('deleteDisk')}
        </form>
      </div>
    )
  }

  function renderPartitionsTab() {
    const create = visualForms.createPartition
    const mount = visualForms.mountPartition
    const format = visualForms.formatPartition
    const remove = visualForms.deletePartition
    const resize = visualForms.resizePartition

    return (
      <div className="action-grid">
        <form
          className="action-card wide"
          onSubmit={(event) =>
            executeVisualAction(
              event,
              'createPartition',
              `fdisk -size=${commandValue(create.size)} -unit=${commandValue(
                create.unit,
              )} -path=${commandValue(create.path)} -name=${commandValue(
                create.name,
              )} -type=${commandValue(create.type)} -fit=${commandValue(create.fit)}`,
              { refreshExplorer: false },
            )
          }
        >
          <h2>Crear particion</h2>
          <div className="form-grid">
            {renderTextField('createPartition', 'size', 'Size', { type: 'number', min: '1' })}
            {renderSelectField('createPartition', 'unit', 'Unit', [
              { label: 'B', value: 'B' },
              { label: 'K', value: 'K' },
              { label: 'M', value: 'M' },
            ])}
            {renderTextField('createPartition', 'path', 'Path')}
            {renderTextField('createPartition', 'name', 'Name')}
            {renderSelectField('createPartition', 'type', 'Type', [
              { label: 'Primaria', value: 'P' },
              { label: 'Extendida', value: 'E' },
              { label: 'Logica', value: 'L' },
            ])}
            {renderSelectField('createPartition', 'fit', 'Fit', [
              { label: 'First Fit', value: 'FF' },
              { label: 'Best Fit', value: 'BF' },
              { label: 'Worst Fit', value: 'WF' },
            ])}
          </div>
          <div className="actions-row">{renderActionButton('createPartition', 'Crear particion')}</div>
          {actionResult('createPartition')}
        </form>

        <form
          className="action-card"
          onSubmit={(event) =>
            executeVisualAction(
              event,
              'mountPartition',
              `mount -path=${commandValue(mount.path)} -name=${commandValue(mount.name)}`,
              { refreshExplorer: false },
            )
          }
        >
          <h2>Montar particion</h2>
          <div className="form-grid single">
            {renderTextField('mountPartition', 'path', 'Path')}
            {renderTextField('mountPartition', 'name', 'Name')}
          </div>
          <div className="actions-row">{renderActionButton('mountPartition', 'Montar')}</div>
          {actionResult('mountPartition')}
        </form>

        <form
          className="action-card"
          onSubmit={(event) =>
            executeVisualAction(
              event,
              'formatPartition',
              `mkfs -id=${commandValue(format.id)} -type=${commandValue(format.type)}`,
            )
          }
        >
          <h2>Formatear particion</h2>
          <div className="form-grid single">
            {renderTextField('formatPartition', 'id', 'ID')}
            {renderSelectField('formatPartition', 'type', 'Type', [
              { label: 'Full', value: 'full' },
              { label: 'Fast', value: 'fast' },
            ])}
          </div>
          <div className="actions-row">{renderActionButton('formatPartition', 'Formatear')}</div>
          {actionResult('formatPartition')}
        </form>

        <form
          className="action-card"
          onSubmit={(event) =>
            executeVisualAction(
              event,
              'deletePartition',
              `fdisk -delete=${commandValue(remove.deleteMode)} -path=${commandValue(
                remove.path,
              )} -name=${commandValue(remove.name)}`,
              { refreshExplorer: false },
            )
          }
        >
          <h2>Eliminar particion</h2>
          <div className="form-grid single">
            {renderTextField('deletePartition', 'path', 'Path')}
            {renderTextField('deletePartition', 'name', 'Name')}
            {renderSelectField('deletePartition', 'deleteMode', 'Delete', [
              { label: 'Fast', value: 'fast' },
              { label: 'Full', value: 'full' },
            ])}
          </div>
          <div className="actions-row">{renderActionButton('deletePartition', 'Eliminar')}</div>
          {actionResult('deletePartition')}
        </form>

        <form
          className="action-card"
          onSubmit={(event) =>
            executeVisualAction(
              event,
              'resizePartition',
              `fdisk -add=${commandValue(resize.add)} -unit=${commandValue(
                resize.unit,
              )} -path=${commandValue(resize.path)} -name=${commandValue(resize.name)}`,
              { refreshExplorer: false },
            )
          }
        >
          <h2>Agregar o quitar espacio</h2>
          <div className="form-grid single">
            {renderTextField('resizePartition', 'path', 'Path')}
            {renderTextField('resizePartition', 'name', 'Name')}
            {renderTextField('resizePartition', 'add', 'Add', { type: 'number' })}
            {renderSelectField('resizePartition', 'unit', 'Unit', [
              { label: 'B', value: 'B' },
              { label: 'K', value: 'K' },
              { label: 'M', value: 'M' },
            ])}
          </div>
          <div className="actions-row">{renderActionButton('resizePartition', 'Aplicar')}</div>
          {actionResult('resizePartition')}
        </form>
      </div>
    )
  }

  function renderFilesTab() {
    const folder = visualForms.createFolder
    const file = visualForms.createFile
    const rename = visualForms.renameEntry
    const edit = visualForms.editFile
    const remove = visualForms.removeEntry
    const copy = visualForms.copyEntry
    const move = visualForms.moveEntry
    const fileCommand = file.content.trim()
      ? `mkfile -path=${commandValue(file.path)} -cont=${commandValue(file.content)}`
      : `mkfile -path=${commandValue(file.path)} -size=${commandValue(file.size)}`

    return (
      <div className="action-grid">
        <form
          className="action-card"
          onSubmit={(event) =>
            executeVisualAction(
              event,
              'createFolder',
              `mkdir -path=${commandValue(folder.path)}${folder.recursive ? ' -p' : ''}`,
            )
          }
        >
          <h2>Crear carpeta</h2>
          <div className="form-grid single">
            {renderTextField('createFolder', 'path', 'Path')}
            <label className="checkbox-field">
              <input
                type="checkbox"
                checked={folder.recursive}
                onChange={(event) =>
                  updateVisualForm('createFolder', 'recursive', event.target.checked)
                }
              />
              <span>Crear padres con -p</span>
            </label>
          </div>
          <div className="actions-row">{renderActionButton('createFolder', 'Crear carpeta')}</div>
          {actionResult('createFolder')}
        </form>

        <form
          className="action-card wide"
          onSubmit={(event) => executeVisualAction(event, 'createFile', fileCommand)}
        >
          <h2>Crear archivo</h2>
          <div className="form-grid">
            {renderTextField('createFile', 'path', 'Path')}
            {renderTextField('createFile', 'size', 'Size', { type: 'number', min: '0' })}
            <label className="span-2">
              <span>Contenido</span>
              <textarea
                value={file.content}
                onChange={(event) => updateVisualForm('createFile', 'content', event.target.value)}
                placeholder="Si escribes contenido se usara -cont; si queda vacio se usara -size."
              />
            </label>
          </div>
          <div className="actions-row">{renderActionButton('createFile', 'Crear archivo')}</div>
          {actionResult('createFile')}
        </form>

        <form
          className="action-card"
          onSubmit={(event) =>
            executeVisualAction(
              event,
              'renameEntry',
              `rename -path=${commandValue(rename.path)} -name=${commandValue(rename.name)}`,
            )
          }
        >
          <h2>Renombrar</h2>
          <div className="form-grid single">
            {renderTextField('renameEntry', 'path', 'Path')}
            {renderTextField('renameEntry', 'name', 'Nuevo nombre')}
          </div>
          <div className="actions-row">{renderActionButton('renameEntry', 'Renombrar')}</div>
          {actionResult('renameEntry')}
        </form>

        <form
          className="action-card"
          onSubmit={(event) =>
            executeVisualAction(
              event,
              'editFile',
              `edit -path=${commandValue(edit.path)} -contenido=${commandValue(edit.contentPath)}`,
            )
          }
        >
          <h2>Editar archivo</h2>
          <div className="form-grid single">
            {renderTextField('editFile', 'path', 'Path EXT2')}
            {renderTextField('editFile', 'contentPath', 'Archivo SO para -contenido')}
          </div>
          <div className="actions-row">{renderActionButton('editFile', 'Editar')}</div>
          {actionResult('editFile')}
        </form>

        <form
          className="action-card"
          onSubmit={(event) =>
            executeVisualAction(
              event,
              'removeEntry',
              `remove -path=${commandValue(remove.path)}`,
            )
          }
        >
          <h2>Eliminar</h2>
          <div className="form-grid single">
            {renderTextField('removeEntry', 'path', 'Path')}
          </div>
          <div className="actions-row">{renderActionButton('removeEntry', 'Eliminar')}</div>
          {actionResult('removeEntry')}
        </form>

        <form
          className="action-card"
          onSubmit={(event) =>
            executeVisualAction(
              event,
              'copyEntry',
              `copy -path=${commandValue(copy.path)} -destino=${commandValue(copy.destino)}`,
            )
          }
        >
          <h2>Copiar</h2>
          <div className="form-grid single">
            {renderTextField('copyEntry', 'path', 'Path')}
            {renderTextField('copyEntry', 'destino', 'Destino')}
          </div>
          <div className="actions-row">{renderActionButton('copyEntry', 'Copiar')}</div>
          {actionResult('copyEntry')}
        </form>

        <form
          className="action-card"
          onSubmit={(event) =>
            executeVisualAction(
              event,
              'moveEntry',
              `move -path=${commandValue(move.path)} -destino=${commandValue(move.destino)}`,
            )
          }
        >
          <h2>Mover</h2>
          <div className="form-grid single">
            {renderTextField('moveEntry', 'path', 'Path')}
            {renderTextField('moveEntry', 'destino', 'Destino')}
          </div>
          <div className="actions-row">{renderActionButton('moveEntry', 'Mover')}</div>
          {actionResult('moveEntry')}
        </form>
      </div>
    )
  }

  return (
    <main className="app-shell">
      <section className="explorer-window">
        <header className="window-titlebar">
          <div>
            <p className="eyebrow">MIA Proyecto Fase 2</p>
            <h1>Explorador EXT2</h1>
          </div>
          <div className="session-summary">
            <span>{loadingSession ? 'Consultando sesion...' : 'Usuario conectado'}</span>
            <strong>
              {session.logged
                ? `${session.user} | ${session.group} | ${session.partitionID}`
                : 'Sin sesion'}
            </strong>
          </div>
        </header>

        <div className="window-body">
          <aside className="sidebar">
            <section className="sidebar-section">
              <div className="sidebar-heading">
                <h2>Montajes</h2>
                <button
                  className="icon-button"
                  type="button"
                  onClick={fetchMounts}
                  title="Actualizar montajes"
                >
                  ↻
                </button>
              </div>

              {mountsError && <div className="inline-error">{mountsError}</div>}

              <div className="mount-list">
                {mounts.length === 0 && !mountsError && (
                  <div className="empty-state compact">
                    {loadingMounts ? 'Cargando montajes...' : 'No hay particiones montadas.'}
                  </div>
                )}

                {mounts.map((mount) => (
                  <button
                    type="button"
                    key={mount.id}
                    className={mount.id === selectedMountId ? 'mount-item active' : 'mount-item'}
                    onClick={() => selectMount(mount.id)}
                  >
                    <span className="drive-icon">▣</span>
                    <span className="mount-name">Disco {mount.id}</span>
                    <span className="mount-path">{mount.name || mount.path}</span>
                  </button>
                ))}
              </div>
            </section>

            <section className="sidebar-section login-section">
              <div className="sidebar-heading">
                <h2>Login</h2>
                {loadingSession && <span className="mini-status">...</span>}
              </div>

              {sessionError && <div className="inline-error">{sessionError}</div>}

              {session.logged ? (
                <div className="session-card">
                  <span className="session-label">Conectado como</span>
                  <strong>{session.user}</strong>
                  <span>{session.group} | {session.partitionID}</span>
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
          </aside>

          <section className="content-column">
            <section className="explorer-panel">
              <div className="command-bar">
                <button
                  className="icon-button"
                  type="button"
                  disabled={!selectedMountId || pathHistory.length === 0}
                  onClick={goBack}
                  title="Atras"
                >
                  ←
                </button>
                <button
                  className="icon-button"
                  type="button"
                  disabled={!selectedMountId || currentPath === '/'}
                  onClick={() => navigateTo(parentPath(currentPath))}
                  title="Subir carpeta"
                >
                  ↑
                </button>
                <button
                  className="icon-button"
                  type="button"
                  disabled={!selectedMountId}
                  onClick={() => loadTree(selectedMountId, currentPath)}
                  title="Actualizar"
                >
                  ↻
                </button>

                <nav className="breadcrumbs" aria-label="Ruta actual">
                  {breadcrumbs.map((crumb, index) => (
                    <span className="breadcrumb-item" key={`${crumb.label}-${index}`}>
                      {crumb.path ? (
                        <button
                          type="button"
                          onClick={() => navigateTo(crumb.path)}
                          disabled={!selectedMountId || crumb.path === currentPath}
                        >
                          {crumb.label}
                        </button>
                      ) : (
                        <span>{crumb.label}</span>
                      )}
                      {index < breadcrumbs.length - 1 && <span className="breadcrumb-separator">&gt;</span>}
                    </span>
                  ))}
                </nav>
              </div>

              {treeError && <div className="inline-error">{treeError}</div>}

              <div className="explorer-grid">
                <div className="file-table-wrap">
                  <table className="file-table">
                    <thead>
                      <tr>
                        <th>Nombre</th>
                        <th>Tipo</th>
                        <th>Inodo</th>
                      </tr>
                    </thead>
                    <tbody>
                      {loadingTree && (
                        <tr>
                          <td colSpan="3" className="table-state">Cargando carpeta...</td>
                        </tr>
                      )}

                      {!loadingTree && items.length === 0 && !treeError && (
                        <tr>
                          <td colSpan="3" className="table-state">La carpeta esta vacia.</td>
                        </tr>
                      )}

                      {!loadingTree &&
                        items.map((item) => (
                          <tr
                            key={`${item.inode}-${item.name}`}
                            className="file-row"
                            onClick={() => handleItemClick(item)}
                          >
                            <td>
                              <span className="file-name">
                                <span className="file-icon" aria-hidden="true">
                                  {item.type === 'folder' ? '📁' : '📄'}
                                </span>
                                {item.name}
                              </span>
                            </td>
                            <td>{item.type === 'folder' ? 'Carpeta' : 'Archivo'}</td>
                            <td>{item.inode}</td>
                          </tr>
                        ))}
                    </tbody>
                  </table>
                </div>

                <aside className="file-preview">
                  <div className="preview-heading">
                    <h2>Vista previa</h2>
                    <span>{loadingFile ? 'Leyendo...' : selectedFilePath || 'Sin archivo'}</span>
                  </div>

                  {fileError && <div className="inline-error">{fileError}</div>}

                  <pre className="content-box">
                    {fileContent || 'Selecciona un archivo para ver su contenido.'}
                  </pre>
                </aside>
              </div>
            </section>

            <section className="utility-panel">
              <div className="tab-bar">
                <button
                  type="button"
                  className={activeUtilityTab === 'disks' ? 'tab-button active' : 'tab-button'}
                  onClick={() => setActiveUtilityTab('disks')}
                >
                  Discos
                </button>
                <button
                  type="button"
                  className={activeUtilityTab === 'partitions' ? 'tab-button active' : 'tab-button'}
                  onClick={() => setActiveUtilityTab('partitions')}
                >
                  Particiones
                </button>
                <button
                  type="button"
                  className={activeUtilityTab === 'files' ? 'tab-button active' : 'tab-button'}
                  onClick={() => setActiveUtilityTab('files')}
                >
                  Archivos
                </button>
                <button
                  type="button"
                  className={activeUtilityTab === 'console' ? 'tab-button active' : 'tab-button'}
                  onClick={() => setActiveUtilityTab('console')}
                >
                  Consola
                </button>
                <button
                  type="button"
                  className={activeUtilityTab === 'reports' ? 'tab-button active' : 'tab-button'}
                  onClick={() => setActiveUtilityTab('reports')}
                >
                  Reportes
                </button>
              </div>

              {activeUtilityTab === 'disks' && (
                <div className="tab-content">{renderDisksTab()}</div>
              )}

              {activeUtilityTab === 'partitions' && (
                <div className="tab-content">{renderPartitionsTab()}</div>
              )}

              {activeUtilityTab === 'files' && (
                <div className="tab-content">{renderFilesTab()}</div>
              )}

              {activeUtilityTab === 'console' && (
                <div className="tab-content console-tab">
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
                </div>
              )}

              {activeUtilityTab === 'reports' && (
                <div className="tab-content">
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
                      <select
                        value={reportMountId}
                        onChange={(event) => setReportMountId(event.target.value)}
                      >
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
                        <h2>Vista previa</h2>
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
                </div>
              )}
            </section>
          </section>
        </div>
      </section>
    </main>
  )
}

export default App
