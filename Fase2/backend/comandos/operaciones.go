package comandos

import (
	"fmt"
	"os"
	"strings"

	"MIA_P1_202243063/estructuras"
	"MIA_P1_202243063/utils"
)

func EjecutarRENAME(params map[string]string) error { // esta funcion renombra archivos o carpetas dentro de ext2
	sesion, file, sb, err := abrirSesionExt2()
	if err != nil {
		return err
	}
	defer file.Close()

	path := params["path"]
	nuevoNombre := strings.TrimSpace(params["name"])
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path debe ser una ruta absoluta")
	}

	parentParts, oldName, err := separarRutaArchivo(path)
	if err != nil {
		return err
	}

	if len(parentParts) == 0 && oldName == "" {
		return fmt.Errorf("no se puede renombrar la raiz")
	}

	if nuevoNombre == "" {
		return fmt.Errorf("name es obligatorio")
	}

	if len(nuevoNombre) > 12 {
		return fmt.Errorf("name no debe superar 12 caracteres")
	}

	parentIndex, err := obtenerOCrearRutaPadre(file, &sb, parentParts, false, sesion.Usuario)
	if err != nil {
		return err
	}

	parentInode, err := leerInodoPorIndice(file, sb, parentIndex)
	if err != nil {
		return err
	}

	if err := validarPermisoInodo(parentInode, sesion.Usuario, permisoEscritura, "renombrar dentro de la carpeta padre"); err != nil {
		return err
	}

	targetIndex, exists, err := buscarEntradaEnCarpeta(file, sb, parentIndex, oldName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no existe la ruta %s", path)
	}

	targetInode, err := leerInodoPorIndice(file, sb, targetIndex)
	if err != nil {
		return err
	}
	if err := validarPermisoInodo(targetInode, sesion.Usuario, permisoEscritura, "renombrar el elemento"); err != nil {
		return err
	}

	if _, exists, err := buscarEntradaEnCarpeta(file, sb, parentIndex, nuevoNombre); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("ya existe un elemento llamado %s en la carpeta padre", nuevoNombre)
	}

	if err := renombrarEntradaEnCarpeta(file, sb, parentIndex, oldName, nuevoNombre); err != nil {
		return err
	}

	fmt.Printf("elemento renombrado correctamente: %s -> %s\n", path, nuevoNombre)
	return nil
}

func EjecutarEDIT(params map[string]string) error { // esta funcion reemplaza el contenido de un archivo ext2 desde un archivo del sistema operativo
	sesion, file, sb, err := abrirSesionExt2()
	if err != nil {
		return err
	}
	defer file.Close()

	path := params["path"]
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path debe ser una ruta absoluta")
	}

	contenidoPath := strings.TrimSpace(params["contenido"])
	if contenidoPath == "" {
		return fmt.Errorf("contenido es obligatorio")
	}

	contenido, err := os.ReadFile(contenidoPath)
	if err != nil {
		return fmt.Errorf("no se pudo leer archivo contenido: %w", err)
	}

	inodeIndex, err := buscarInodoPorRutaConUsuario(file, sb, path, sesion.Usuario)
	if err != nil {
		return err
	}

	inode, err := leerInodoPorIndice(file, sb, inodeIndex)
	if err != nil {
		return err
	}

	if inode.Type != estructuras.FileBlockType {
		return fmt.Errorf("%s no es un archivo", path)
	}

	if err := validarPermisoInodo(inode, sesion.Usuario, permisoEscritura, "editar el archivo"); err != nil {
		return err
	}

	if err := escribirContenidoArchivo(file, &sb, &inode, contenido, sesion.Montada.Partition.Fit); err != nil {
		return err
	}

	if err := escribirInodoPorIndice(file, sb, inodeIndex, inode); err != nil {
		return err
	}

	if err := utils.WriteStructAt(file, int64(sesion.Montada.Partition.Start), &sb); err != nil {
		return err
	}

	fmt.Printf("archivo editado correctamente: %s\n", path)
	return nil
}

func EjecutarREMOVE(params map[string]string) error { // esta funcion elimina archivos o carpetas recursivamente
	sesion, file, sb, err := abrirSesionExt2()
	if err != nil {
		return err
	}
	defer file.Close()

	path := params["path"]
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path debe ser una ruta absoluta")
	}

	parentParts, name, err := separarRutaArchivo(path)
	if err != nil {
		return err
	}
	if len(parentParts) == 0 && name == "" {
		return fmt.Errorf("no se puede eliminar la raiz")
	}

	parentIndex, err := obtenerOCrearRutaPadre(file, &sb, parentParts, false, sesion.Usuario)
	if err != nil {
		return err
	}

	parentInode, err := leerInodoPorIndice(file, sb, parentIndex)
	if err != nil {
		return err
	}
	if err := validarPermisoInodo(parentInode, sesion.Usuario, permisoEscritura, "eliminar dentro de la carpeta padre"); err != nil {
		return err
	}

	targetIndex, exists, err := buscarEntradaEnCarpeta(file, sb, parentIndex, name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no existe la ruta %s", path)
	}

	if err := validarEliminacionRecursiva(file, sb, targetIndex, sesion.Usuario); err != nil {
		return err
	}

	if err := eliminarInodoRecursivo(file, &sb, targetIndex); err != nil {
		return err
	}

	if err := eliminarEntradaEnCarpeta(file, sb, parentIndex, name); err != nil {
		return err
	}

	if err := utils.WriteStructAt(file, int64(sesion.Montada.Partition.Start), &sb); err != nil {
		return err
	}

	fmt.Printf("elemento eliminado correctamente: %s\n", path)
	return nil
}

func EjecutarCOPY(params map[string]string) error { // esta funcion copia archivos o carpetas dentro de ext2
	sesion, file, sb, err := abrirSesionExt2()
	if err != nil {
		return err
	}
	defer file.Close()

	path := strings.TrimSpace(params["path"])
	destino := strings.TrimSpace(params["destino"])
	if !strings.HasPrefix(path, "/") || !strings.HasPrefix(destino, "/") {
		return fmt.Errorf("path y destino deben ser rutas absolutas")
	}

	_, sourceName, err := separarRutaArchivo(path)
	if err != nil {
		return err
	}
	if sourceName == "" {
		return fmt.Errorf("no se puede copiar la raiz")
	}

	sourceIndex, err := buscarInodoPorRutaConUsuario(file, sb, path, sesion.Usuario)
	if err != nil {
		return err
	}
	sourceInode, err := leerInodoPorIndice(file, sb, sourceIndex)
	if err != nil {
		return err
	}
	if err := validarPermisoInodo(sourceInode, sesion.Usuario, permisoLectura, "copiar el origen"); err != nil {
		return err
	}
	if sourceInode.Type == estructuras.FolderBlockType && rutaDentroDeSiMisma(path, destino) {
		return fmt.Errorf("no se puede copiar una carpeta dentro de si misma")
	}

	destIndex, err := buscarInodoPorRutaConUsuario(file, sb, destino, sesion.Usuario)
	if err != nil {
		return err
	}
	destInode, err := leerInodoPorIndice(file, sb, destIndex)
	if err != nil {
		return err
	}
	if destInode.Type != estructuras.FolderBlockType {
		return fmt.Errorf("destino debe ser una carpeta")
	}
	if err := validarPermisoInodo(destInode, sesion.Usuario, permisoEscritura, "copiar dentro del destino"); err != nil {
		return err
	}
	if _, exists, err := buscarEntradaEnCarpeta(file, sb, destIndex, sourceName); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("ya existe un elemento llamado %s en el destino", sourceName)
	}

	if err := copiarInodoRecursivo(file, &sb, sourceIndex, destIndex, sourceName, sesion.Usuario, sesion.Montada.Partition.Fit, false); err != nil {
		return err
	}
	if err := utils.WriteStructAt(file, int64(sesion.Montada.Partition.Start), &sb); err != nil {
		return err
	}

	fmt.Printf("elemento copiado correctamente: %s -> %s\n", path, destino)
	return nil
}

func EjecutarMOVE(params map[string]string) error { // esta funcion mueve archivos o carpetas cambiando referencias de carpeta
	sesion, file, sb, err := abrirSesionExt2()
	if err != nil {
		return err
	}
	defer file.Close()

	path := strings.TrimSpace(params["path"])
	destino := strings.TrimSpace(params["destino"])
	if !strings.HasPrefix(path, "/") || !strings.HasPrefix(destino, "/") {
		return fmt.Errorf("path y destino deben ser rutas absolutas")
	}

	parentParts, sourceName, err := separarRutaArchivo(path)
	if err != nil {
		return err
	}
	if sourceName == "" {
		return fmt.Errorf("no se puede mover la raiz")
	}

	sourceIndex, err := buscarInodoPorRutaConUsuario(file, sb, path, sesion.Usuario)
	if err != nil {
		return err
	}
	sourceInode, err := leerInodoPorIndice(file, sb, sourceIndex)
	if err != nil {
		return err
	}
	if err := validarPermisoInodo(sourceInode, sesion.Usuario, permisoEscritura, "mover el origen"); err != nil {
		return err
	}

	sourceParentIndex, err := obtenerOCrearRutaPadre(file, &sb, parentParts, false, sesion.Usuario)
	if err != nil {
		return err
	}
	sourceParentInode, err := leerInodoPorIndice(file, sb, sourceParentIndex)
	if err != nil {
		return err
	}
	if err := validarPermisoInodo(sourceParentInode, sesion.Usuario, permisoEscritura, "quitar el origen de su carpeta padre"); err != nil {
		return err
	}

	destIndex, err := buscarInodoPorRutaConUsuario(file, sb, destino, sesion.Usuario)
	if err != nil {
		return err
	}
	destInode, err := leerInodoPorIndice(file, sb, destIndex)
	if err != nil {
		return err
	}
	if destInode.Type != estructuras.FolderBlockType {
		return fmt.Errorf("destino debe ser una carpeta")
	}
	if err := validarPermisoInodo(destInode, sesion.Usuario, permisoEscritura, "mover dentro del destino"); err != nil {
		return err
	}

	if sourceInode.Type == estructuras.FolderBlockType && rutaDentroDeSiMisma(path, destino) {
		return fmt.Errorf("no se puede mover una carpeta dentro de si misma")
	}
	if _, exists, err := buscarEntradaEnCarpeta(file, sb, destIndex, sourceName); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("ya existe un elemento llamado %s en el destino", sourceName)
	}

	if err := eliminarEntradaEnCarpeta(file, sb, sourceParentIndex, sourceName); err != nil {
		return err
	}
	if err := escribirEntradaEnCarpeta(file, &sb, destIndex, sourceName, sourceIndex); err != nil {
		return err
	}
	if err := utils.WriteStructAt(file, int64(sesion.Montada.Partition.Start), &sb); err != nil {
		return err
	}

	fmt.Printf("elemento movido correctamente: %s -> %s\n", path, destino)
	return nil
}

func abrirSesionExt2() (SesionSistema, *os.File, estructuras.SuperBlock, error) {
	sesion, activa := obtenerSesionActual()
	if !activa {
		return SesionSistema{}, nil, estructuras.SuperBlock{}, fmt.Errorf("no existe una sesion activa")
	}

	file, err := os.OpenFile(sesion.Montada.Path, os.O_RDWR, 0666)
	if err != nil {
		return SesionSistema{}, nil, estructuras.SuperBlock{}, fmt.Errorf("no se pudo abrir el disco: %w", err)
	}

	var sb estructuras.SuperBlock
	if err := utils.ReadStructAt(file, int64(sesion.Montada.Partition.Start), &sb); err != nil {
		file.Close()
		return SesionSistema{}, nil, estructuras.SuperBlock{}, err
	}

	if sb.Magic != estructuras.Ext2Magic {
		file.Close()
		return SesionSistema{}, nil, estructuras.SuperBlock{}, fmt.Errorf("la particion no esta formateada como ext2")
	}

	return sesion, file, sb, nil
}

func renombrarEntradaEnCarpeta(file *os.File, sb estructuras.SuperBlock, parentIndex int32, oldName string, newName string) error {
	parent, err := leerInodoPorIndice(file, sb, parentIndex)
	if err != nil {
		return err
	}

	for _, blockIndex := range parent.Block[:12] {
		if blockIndex == -1 {
			continue
		}

		block, err := leerBloqueCarpeta(file, sb, blockIndex)
		if err != nil {
			return err
		}

		for contentIndex := range block.Content {
			if block.Content[contentIndex].Inode != -1 && utils.BytesToString(block.Content[contentIndex].Name[:]) == oldName {
				block.Content[contentIndex].Name = utils.StringToBytes12(newName)
				return escribirBloqueCarpeta(file, sb, blockIndex, block)
			}
		}
	}

	return fmt.Errorf("no se encontro la entrada %s en la carpeta padre", oldName)
}

func eliminarEntradaEnCarpeta(file *os.File, sb estructuras.SuperBlock, parentIndex int32, name string) error {
	parent, err := leerInodoPorIndice(file, sb, parentIndex)
	if err != nil {
		return err
	}

	for _, blockIndex := range parent.Block[:12] {
		if blockIndex == -1 {
			continue
		}

		block, err := leerBloqueCarpeta(file, sb, blockIndex)
		if err != nil {
			return err
		}

		for contentIndex := range block.Content {
			if block.Content[contentIndex].Inode != -1 && utils.BytesToString(block.Content[contentIndex].Name[:]) == name {
				block.Content[contentIndex] = estructuras.Content{Name: [12]byte{}, Inode: -1}
				return escribirBloqueCarpeta(file, sb, blockIndex, block)
			}
		}
	}

	return fmt.Errorf("no se encontro la entrada %s en la carpeta padre", name)
}

func validarEliminacionRecursiva(file *os.File, sb estructuras.SuperBlock, inodeIndex int32, usuario UsuarioSistema) error {
	inode, err := leerInodoPorIndice(file, sb, inodeIndex)
	if err != nil {
		return err
	}

	if err := validarPermisoInodo(inode, usuario, permisoEscritura, "eliminar el elemento"); err != nil {
		return err
	}

	if inode.Type != estructuras.FolderBlockType {
		return nil
	}

	entradas, err := entradasEliminablesCarpeta(file, sb, inode)
	if err != nil {
		return err
	}

	for _, entrada := range entradas {
		if err := validarEliminacionRecursiva(file, sb, entrada.Inode, usuario); err != nil {
			return err
		}
	}

	return nil
}

func eliminarInodoRecursivo(file *os.File, sb *estructuras.SuperBlock, inodeIndex int32) error {
	inode, err := leerInodoPorIndice(file, *sb, inodeIndex)
	if err != nil {
		return err
	}

	if inode.Type == estructuras.FileBlockType {
		if err := liberarBloquesArchivo(file, sb, &inode); err != nil {
			return err
		}
		if err := marcarInodo(file, sb, inodeIndex, false); err != nil {
			return err
		}
		return escribirInodoPorIndice(file, *sb, inodeIndex, estructuras.Inode{})
	}

	entradas, err := entradasEliminablesCarpeta(file, *sb, inode)
	if err != nil {
		return err
	}
	for _, entrada := range entradas {
		if err := eliminarInodoRecursivo(file, sb, entrada.Inode); err != nil {
			return err
		}
	}

	for blockIndex := 0; blockIndex < 12; blockIndex++ {
		if inode.Block[blockIndex] == -1 {
			continue
		}
		if err := marcarBloque(file, sb, inode.Block[blockIndex], false); err != nil {
			return err
		}
		inode.Block[blockIndex] = -1
	}

	if err := marcarInodo(file, sb, inodeIndex, false); err != nil {
		return err
	}
	return escribirInodoPorIndice(file, *sb, inodeIndex, estructuras.Inode{})
}

func entradasEliminablesCarpeta(file *os.File, sb estructuras.SuperBlock, inode estructuras.Inode) ([]entradaReporte, error) {
	var entradas []entradaReporte

	for _, blockIndex := range inode.Block[:12] {
		if blockIndex == -1 {
			continue
		}

		block, err := leerBloqueCarpeta(file, sb, blockIndex)
		if err != nil {
			return nil, err
		}

		for _, content := range block.Content {
			name := utils.BytesToString(content.Name[:])
			if content.Inode == -1 || name == "" || name == "." || name == ".." {
				continue
			}
			entradas = append(entradas, entradaReporte{Nombre: name, Inode: content.Inode})
		}
	}

	return entradas, nil
}

func copiarInodoRecursivo(file *os.File, sb *estructuras.SuperBlock, sourceIndex int32, destParentIndex int32, name string, usuario UsuarioSistema, fit byte, omitirSinPermiso bool) error {
	sourceInode, err := leerInodoPorIndice(file, *sb, sourceIndex)
	if err != nil {
		return err
	}

	if err := validarPermisoInodo(sourceInode, usuario, permisoLectura, "copiar el elemento"); err != nil {
		if omitirSinPermiso {
			fmt.Printf("copy: se omitio %s por falta de permiso de lectura\n", name)
			return nil
		}
		return err
	}

	if sourceInode.Type == estructuras.FileBlockType {
		contenido, err := leerContenidoArchivo(file, *sb, sourceInode)
		if err != nil {
			return err
		}
		return crearArchivo(file, sb, destParentIndex, name, []byte(contenido), usuario.UID, usuario.GID, fit)
	}

	newFolderIndex, err := crearCarpeta(file, sb, destParentIndex, name, usuario.UID, usuario.GID)
	if err != nil {
		return err
	}

	entradas, err := entradasEliminablesCarpeta(file, *sb, sourceInode)
	if err != nil {
		return err
	}
	for _, entrada := range entradas {
		childInode, err := leerInodoPorIndice(file, *sb, entrada.Inode)
		if err != nil {
			return err
		}
		if err := validarPermisoInodo(childInode, usuario, permisoLectura, "copiar el hijo"); err != nil {
			fmt.Printf("copy: se omitio %s por falta de permiso de lectura\n", entrada.Nombre)
			continue
		}
		if _, exists, err := buscarEntradaEnCarpeta(file, *sb, newFolderIndex, entrada.Nombre); err != nil {
			return err
		} else if exists {
			fmt.Printf("copy: se omitio %s porque ya existe en el destino\n", entrada.Nombre)
			continue
		}
		if err := copiarInodoRecursivo(file, sb, entrada.Inode, newFolderIndex, entrada.Nombre, usuario, fit, true); err != nil {
			return err
		}
	}

	return nil
}

func rutaDentroDeSiMisma(origen string, destino string) bool {
	origen = limpiarRutaOperacion(origen)
	destino = limpiarRutaOperacion(destino)
	return destino == origen || strings.HasPrefix(destino, origen+"/")
}

func limpiarRutaOperacion(path string) string {
	path = "/" + strings.Join(separarRuta(path), "/")
	if path == "" {
		return "/"
	}
	return path
}
