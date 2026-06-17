package comandos

import (
	"fmt"     // uso fmt para errores y salida de cat
	"os"      // uso os para abrir discos y leer archivos externos con -cont
	"sort"    // uso sort para ordenar file1, file2, file3
	"strconv" // uso strconv para convertir size y numero de fileN
	"strings" // uso strings para validar rutas y nombres

	"MIA_P1_202243063/estructuras" // uso estructuras para inodos y bloques de archivo
	"MIA_P1_202243063/utils"       // uso utils para conversiones y lectura binaria
)

func EjecutarMKFILE(params map[string]string, flags map[string]bool) error { // esta funcion crea o sobrescribe archivos dentro de ext2
	sesion, activa := obtenerSesionActual()
	if !activa {
		return fmt.Errorf("no existe una sesion activa")
	}

	path := params["path"] // obtengo la ruta del archivo a crear
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path debe ser una ruta absoluta")
	}

	parentPath, fileName, err := separarRutaArchivo(path)
	if err != nil {
		return err
	}

	if len(fileName) > 12 {
		return fmt.Errorf("el nombre del archivo debe tener maximo 12 caracteres")
	}

	content, err := construirContenidoArchivo(params) // genero o cargo el contenido segun -size y -cont
	if err != nil {
		return err
	}

	file, err := os.OpenFile(sesion.Montada.Path, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el disco: %w", err)
	}
	defer file.Close()

	var sb estructuras.SuperBlock
	if err := utils.ReadStructAt(file, int64(sesion.Montada.Partition.Start), &sb); err != nil {
		return err
	}

	parentIndex, err := obtenerOCrearRutaPadre(file, &sb, parentPath, flags["r"], sesion.Usuario.UID, sesion.Usuario.GID)
	if err != nil {
		return err
	}

	existingIndex, exists, err := buscarEntradaEnCarpeta(file, sb, parentIndex, fileName)
	if err != nil {
		return err
	}

	if exists {
		if err := sobrescribirArchivo(file, &sb, existingIndex, content); err != nil {
			return err
		}
	} else {
		if err := crearArchivo(file, &sb, parentIndex, fileName, content, sesion.Usuario.UID, sesion.Usuario.GID); err != nil {
			return err
		}
	}

	if err := utils.WriteStructAt(file, int64(sesion.Montada.Partition.Start), &sb); err != nil {
		return err
	}

	fmt.Printf("archivo creado correctamente: %s\n", path)
	return nil
}

func EjecutarCAT(params map[string]string) error { // esta funcion muestra el contenido de uno o varios archivos ext2
	sesion, activa := obtenerSesionActual()
	if !activa {
		return fmt.Errorf("no existe una sesion activa")
	}

	files := ordenarParametrosFile(params)
	if len(files) == 0 {
		return fmt.Errorf("cat requiere al menos un parametro -fileN")
	}

	file, err := os.Open(sesion.Montada.Path)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el disco: %w", err)
	}
	defer file.Close()

	var sb estructuras.SuperBlock
	if err := utils.ReadStructAt(file, int64(sesion.Montada.Partition.Start), &sb); err != nil {
		return err
	}

	for index, path := range files {
		contenido, err := leerArchivoPorRuta(file, sb, path)
		if err != nil {
			return err
		}

		if index > 0 {
			fmt.Println()
		}
		fmt.Print(contenido)
	}

	if len(files) > 0 {
		fmt.Println()
	}

	return nil
}

func separarRutaArchivo(path string) ([]string, string, error) { // esta funcion separa la ruta padre y el nombre del archivo
	partes := separarRuta(path)
	if len(partes) == 0 {
		return nil, "", fmt.Errorf("path debe incluir nombre de archivo")
	}

	return partes[:len(partes)-1], partes[len(partes)-1], nil
}

func construirContenidoArchivo(params map[string]string) ([]byte, error) { // esta funcion decide el contenido de mkfile
	if contPath := params["cont"]; contPath != "" {
		content, err := os.ReadFile(contPath)
		if err != nil {
			return nil, fmt.Errorf("no se pudo leer archivo cont: %w", err)
		}
		return content, nil
	}

	size := 0
	if params["size"] != "" {
		value, err := strconv.Atoi(params["size"])
		if err != nil {
			return nil, fmt.Errorf("size debe ser un numero entero")
		}

		if value < 0 {
			return nil, fmt.Errorf("size no puede ser negativo")
		}
		size = value
	}

	content := make([]byte, size)
	for index := 0; index < size; index++ {
		content[index] = byte('0' + (index % 10))
	}

	return content, nil
}

func obtenerOCrearRutaPadre(file *os.File, sb *estructuras.SuperBlock, partes []string, crear bool, uid int32, gid int32) (int32, error) { // esta funcion ubica o crea carpetas padre
	current := int32(0)

	for _, nombre := range partes {
		child, exists, err := buscarEntradaEnCarpeta(file, *sb, current, nombre)
		if err != nil {
			return 0, err
		}

		if exists {
			current = child
			continue
		}

		if !crear {
			return 0, fmt.Errorf("no existe la carpeta padre %s, usa -r para crear padres", nombre)
		}

		newIndex, err := crearCarpeta(file, sb, current, nombre, uid, gid)
		if err != nil {
			return 0, err
		}
		current = newIndex
	}

	return current, nil
}

func crearArchivo(file *os.File, sb *estructuras.SuperBlock, parentIndex int32, nombre string, contenido []byte, uid int32, gid int32) error { // esta funcion reserva inodo para un archivo nuevo
	inodeIndex, err := buscarInodoLibre(file, *sb)
	if err != nil {
		return err
	}

	if err := marcarInodo(file, sb, inodeIndex, true); err != nil {
		return err
	}

	inode := createInitialInode(uid, gid, 0, estructuras.FileBlockType, "664")
	if err := escribirContenidoArchivo(file, sb, &inode, contenido); err != nil {
		return err
	}

	if err := escribirEntradaEnCarpeta(file, sb, parentIndex, nombre, inodeIndex); err != nil {
		return err
	}

	return escribirInodoPorIndice(file, *sb, inodeIndex, inode)
}

func sobrescribirArchivo(file *os.File, sb *estructuras.SuperBlock, inodeIndex int32, contenido []byte) error { // esta funcion reemplaza contenido si el archivo ya existe
	inode, err := leerInodoPorIndice(file, *sb, inodeIndex)
	if err != nil {
		return err
	}

	if inode.Type != estructuras.FileBlockType {
		return fmt.Errorf("ya existe una carpeta con ese nombre")
	}

	if err := escribirContenidoArchivo(file, sb, &inode, contenido); err != nil {
		return err
	}

	return escribirInodoPorIndice(file, *sb, inodeIndex, inode)
}

func leerArchivoPorRuta(file *os.File, sb estructuras.SuperBlock, path string) (string, error) { // esta funcion lee un archivo usando su ruta absoluta
	if !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("file debe ser una ruta absoluta")
	}

	inodeIndex, err := buscarInodoPorRuta(file, sb, path)
	if err != nil {
		return "", err
	}

	inode, err := leerInodoPorIndice(file, sb, inodeIndex)
	if err != nil {
		return "", err
	}

	if inode.Type != estructuras.FileBlockType {
		return "", fmt.Errorf("%s no es un archivo", path)
	}

	return leerContenidoArchivo(file, sb, inode)
}

func buscarInodoPorRuta(file *os.File, sb estructuras.SuperBlock, path string) (int32, error) { // esta funcion recorre carpetas hasta encontrar un inodo
	partes := separarRuta(path)
	if len(partes) == 0 {
		return 0, nil
	}

	current := int32(0)
	for _, nombre := range partes {
		next, exists, err := buscarEntradaEnCarpeta(file, sb, current, nombre)
		if err != nil {
			return 0, err
		}

		if !exists {
			return 0, fmt.Errorf("no existe la ruta %s", path)
		}

		current = next
	}

	return current, nil
}

func ordenarParametrosFile(params map[string]string) []string { // esta funcion ordena file1, file2, file3 para cat
	type entrada struct {
		numero int
		path   string
	}

	var entradas []entrada
	for key, value := range params {
		if !strings.HasPrefix(key, "file") {
			continue
		}

		numeroTexto := strings.TrimPrefix(key, "file")
		numero, err := strconv.Atoi(numeroTexto)
		if err != nil {
			numero = len(entradas) + 1
		}

		entradas = append(entradas, entrada{numero: numero, path: value})
	}

	sort.Slice(entradas, func(i, j int) bool {
		return entradas[i].numero < entradas[j].numero
	})

	files := make([]string, 0, len(entradas))
	for _, entrada := range entradas {
		files = append(files, entrada.path)
	}

	return files
}
