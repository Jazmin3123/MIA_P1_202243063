package comandos

import (
	"fmt"     // uso fmt para mensajes de error y confirmacion
	"os"      // uso os para abrir el disco donde esta users.txt
	"strconv" // uso strconv para calcular ids nuevos
	"strings" // uso strings para reconstruir users.txt
	"time"    // uso time para actualizar fechas del inodo

	"MIA_P1_202243063/estructuras" // uso estructuras para leer y escribir ext2
	"MIA_P1_202243063/utils"       // uso utils para acceso binario
)

func EjecutarMKGRP(params map[string]string) error { // esta funcion crea un grupo dentro de users.txt
	sesion, err := validarSesionRoot() // valido que exista sesion y que sea root
	if err != nil {
		return err
	}

	nombre := params["name"] // obtengo el nombre del grupo
	if len(nombre) > 10 {
		return fmt.Errorf("name no debe superar 10 caracteres")
	}

	contenido, err := leerUsersTxt(sesion.Montada)
	if err != nil {
		return err
	}

	registros := parsearRegistrosUsers(contenido)
	if existeGrupoActivo(registros, nombre) {
		return fmt.Errorf("el grupo %s ya existe", nombre)
	}

	nuevoID := siguienteIDGrupo(registros)
	registros = append(registros, []string{strconv.Itoa(nuevoID), "G", nombre})

	if err := escribirUsersTxt(sesion.Montada, serializarRegistrosUsers(registros)); err != nil {
		return err
	}

	fmt.Printf("grupo creado correctamente: %s\n", nombre)
	return nil
}

func EjecutarRMGRP(params map[string]string) error { // esta funcion elimina logicamente un grupo en users.txt
	sesion, err := validarSesionRoot()
	if err != nil {
		return err
	}

	nombre := params["name"]
	contenido, err := leerUsersTxt(sesion.Montada)
	if err != nil {
		return err
	}

	registros := parsearRegistrosUsers(contenido)
	encontrado := false

	for index := range registros {
		if esGrupoActivo(registros[index]) && registros[index][2] == nombre {
			registros[index][0] = "0" // id cero significa eliminado
			encontrado = true
			break
		}
	}

	if !encontrado {
		return fmt.Errorf("el grupo %s no existe", nombre)
	}

	if err := escribirUsersTxt(sesion.Montada, serializarRegistrosUsers(registros)); err != nil {
		return err
	}

	fmt.Printf("grupo eliminado correctamente: %s\n", nombre)
	return nil
}

func EjecutarMKUSR(params map[string]string) error { // esta funcion crea un usuario dentro de users.txt
	sesion, err := validarSesionRoot()
	if err != nil {
		return err
	}

	usuario := params["user"]
	password := params["pass"]
	grupo := params["grp"]

	if len(usuario) > 10 || len(password) > 10 || len(grupo) > 10 {
		return fmt.Errorf("user, pass y grp no deben superar 10 caracteres")
	}

	contenido, err := leerUsersTxt(sesion.Montada)
	if err != nil {
		return err
	}

	registros := parsearRegistrosUsers(contenido)
	if !existeGrupoActivo(registros, grupo) {
		return fmt.Errorf("el grupo %s no existe", grupo)
	}

	if existeUsuarioActivo(registros, usuario) {
		return fmt.Errorf("el usuario %s ya existe", usuario)
	}

	nuevoID := siguienteIDUsuario(registros)
	registros = append(registros, []string{strconv.Itoa(nuevoID), "U", grupo, usuario, password})

	if err := escribirUsersTxt(sesion.Montada, serializarRegistrosUsers(registros)); err != nil {
		return err
	}

	fmt.Printf("usuario creado correctamente: %s\n", usuario)
	return nil
}

func EjecutarRMUSR(params map[string]string) error { // esta funcion elimina logicamente un usuario en users.txt
	sesion, err := validarSesionRoot()
	if err != nil {
		return err
	}

	usuario := params["user"]
	contenido, err := leerUsersTxt(sesion.Montada)
	if err != nil {
		return err
	}

	registros := parsearRegistrosUsers(contenido)
	encontrado := false

	for index := range registros {
		if esUsuarioActivo(registros[index]) && registros[index][3] == usuario {
			registros[index][0] = "0" // id cero significa eliminado
			encontrado = true
			break
		}
	}

	if !encontrado {
		return fmt.Errorf("el usuario %s no existe", usuario)
	}

	if err := escribirUsersTxt(sesion.Montada, serializarRegistrosUsers(registros)); err != nil {
		return err
	}

	fmt.Printf("usuario eliminado correctamente: %s\n", usuario)
	return nil
}

func EjecutarCHGRP(params map[string]string) error { // esta funcion cambia el grupo de un usuario
	sesion, err := validarSesionRoot()
	if err != nil {
		return err
	}

	usuario := params["user"]
	grupo := params["grp"]
	contenido, err := leerUsersTxt(sesion.Montada)
	if err != nil {
		return err
	}

	registros := parsearRegistrosUsers(contenido)
	if !existeGrupoActivo(registros, grupo) {
		return fmt.Errorf("el grupo %s no existe", grupo)
	}

	encontrado := false
	for index := range registros {
		if esUsuarioActivo(registros[index]) && registros[index][3] == usuario {
			registros[index][2] = grupo
			encontrado = true
			break
		}
	}

	if !encontrado {
		return fmt.Errorf("el usuario %s no existe", usuario)
	}

	if err := escribirUsersTxt(sesion.Montada, serializarRegistrosUsers(registros)); err != nil {
		return err
	}

	fmt.Printf("grupo del usuario %s actualizado a %s\n", usuario, grupo)
	return nil
}

func validarSesionRoot() (SesionSistema, error) { // esta funcion valida que haya sesion activa de root
	sesion, activa := obtenerSesionActual()
	if !activa {
		return SesionSistema{}, fmt.Errorf("no existe una sesion activa")
	}

	if sesion.Usuario.Usuario != "root" {
		return SesionSistema{}, fmt.Errorf("solo el usuario root puede ejecutar este comando")
	}

	return sesion, nil
}

func parsearRegistrosUsers(contenido string) [][]string { // esta funcion convierte users.txt en registros editables
	var registros [][]string

	for _, linea := range strings.Split(contenido, "\n") {
		linea = strings.TrimSpace(linea)
		if linea == "" {
			continue
		}

		registros = append(registros, separarRegistroUsers(linea))
	}

	return registros
}

func serializarRegistrosUsers(registros [][]string) string { // esta funcion reconstruye el texto de users.txt
	var builder strings.Builder

	for _, registro := range registros {
		builder.WriteString(strings.Join(registro, ","))
		builder.WriteString("\n")
	}

	return builder.String()
}

func esGrupoActivo(registro []string) bool { // esta funcion identifica un registro de grupo activo
	return len(registro) >= 3 && registro[0] != "0" && strings.ToUpper(registro[1]) == "G"
}

func esUsuarioActivo(registro []string) bool { // esta funcion identifica un registro de usuario activo
	return len(registro) >= 5 && registro[0] != "0" && strings.ToUpper(registro[1]) == "U"
}

func existeGrupoActivo(registros [][]string, nombre string) bool { // esta funcion busca un grupo activo por nombre
	for _, registro := range registros {
		if esGrupoActivo(registro) && registro[2] == nombre {
			return true
		}
	}

	return false
}

func existeUsuarioActivo(registros [][]string, usuario string) bool { // esta funcion busca un usuario activo por nombre
	for _, registro := range registros {
		if esUsuarioActivo(registro) && registro[3] == usuario {
			return true
		}
	}

	return false
}

func siguienteIDGrupo(registros [][]string) int { // esta funcion calcula el siguiente id de grupo
	return siguienteIDPorTipo(registros, "G")
}

func siguienteIDUsuario(registros [][]string) int { // esta funcion calcula el siguiente id de usuario
	return siguienteIDPorTipo(registros, "U")
}

func siguienteIDPorTipo(registros [][]string, tipo string) int { // esta funcion busca el mayor id usado para un tipo
	mayor := 0

	for _, registro := range registros {
		if len(registro) < 2 || strings.ToUpper(registro[1]) != tipo {
			continue
		}

		id, err := strconv.Atoi(registro[0])
		if err == nil && id > mayor {
			mayor = id
		}
	}

	return mayor + 1
}

func escribirUsersTxt(mounted MountedPartition, contenido string) error { // esta funcion reescribe /users.txt en la particion ext2
	file, err := os.OpenFile(mounted.Path, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el disco: %w", err)
	}
	defer file.Close()

	var sb estructuras.SuperBlock
	if err := utils.ReadStructAt(file, int64(mounted.Partition.Start), &sb); err != nil {
		return err
	}

	var inode estructuras.Inode
	inodePosition := int64(sb.InodeStart + sb.InodeSize)
	if err := utils.ReadStructAt(file, inodePosition, &inode); err != nil {
		return err
	}

	if err := escribirContenidoArchivo(file, &sb, &inode, []byte(contenido), mounted.Partition.Fit); err != nil {
		return err
	}

	if err := utils.WriteStructAt(file, inodePosition, &inode); err != nil {
		return err
	}

	if err := utils.WriteStructAt(file, int64(mounted.Partition.Start), &sb); err != nil {
		return err
	}

	return nil
}

func escribirContenidoArchivo(file *os.File, sb *estructuras.SuperBlock, inode *estructuras.Inode, contenido []byte, fit byte) error { // esta funcion escribe contenido usando bloques directos
	bloquesNecesarios := bloquesNecesariosParaContenido(len(contenido), int(sb.BlockSize))
	if bloquesNecesarios > 12 {
		return fmt.Errorf("users.txt supera la capacidad soportada de bloques directos")
	}

	if err := ajustarBloquesArchivo(file, sb, inode, bloquesNecesarios, fit); err != nil {
		return err
	}

	for index := 0; index < bloquesNecesarios; index++ {
		var block estructuras.FileBlock
		inicio := index * int(sb.BlockSize)
		fin := inicio + int(sb.BlockSize)
		if fin > len(contenido) {
			fin = len(contenido)
		}

		copy(block.Content[:], contenido[inicio:fin])
		blockPosition := int64(sb.BlockStart + inode.Block[index]*sb.BlockSize)
		if err := utils.WriteStructAt(file, blockPosition, &block); err != nil {
			return err
		}
	}

	inode.Size = int32(len(contenido))
	inode.MTime = time.Now().Unix()
	return nil
}

func bloquesNecesariosParaContenido(size int, blockSize int) int { // esta funcion calcula cuantos bloques necesita un contenido
	if size == 0 {
		return 0
	}

	return (size + blockSize - 1) / blockSize
}

func ajustarBloquesArchivo(file *os.File, sb *estructuras.SuperBlock, inode *estructuras.Inode, necesarios int, fit byte) error { // esta funcion asigna o libera bloques directos
	actuales := bloquesDirectosUsados(*inode)

	for len(actuales) > necesarios {
		ultimo := actuales[len(actuales)-1]
		if err := marcarBloque(file, sb, ultimo, false); err != nil {
			return err
		}

		inode.Block[len(actuales)-1] = -1
		actuales = actuales[:len(actuales)-1]
	}

	if len(actuales) < necesarios {
		faltantes := necesarios - len(actuales)
		libres, err := buscarBloquesContiguosLibres(file, *sb, faltantes, fit)
		if err != nil {
			return err
		}

		for _, libre := range libres {
			if err := marcarBloque(file, sb, libre, true); err != nil {
				return err
			}

			inode.Block[len(actuales)] = libre
			actuales = append(actuales, libre)
		}
	}

	return nil
}

func bloquesDirectosUsados(inode estructuras.Inode) []int32 { // esta funcion obtiene los bloques directos ocupados del inodo
	var usados []int32

	for index := 0; index < 12; index++ {
		if inode.Block[index] != -1 {
			usados = append(usados, inode.Block[index])
		}
	}

	return usados
}

func buscarBloqueLibre(file *os.File, sb estructuras.SuperBlock) (int32, error) { // esta funcion busca el primer bloque libre en el bitmap
	bitmap := make([]byte, sb.BlocksCount)
	if _, err := file.ReadAt(bitmap, int64(sb.BmBlockStart)); err != nil {
		return 0, fmt.Errorf("no se pudo leer bitmap de bloques: %w", err)
	}

	for index, value := range bitmap {
		if value == 0 || value == '0' {
			return int32(index), nil
		}
	}

	return 0, fmt.Errorf("no hay bloques libres")
}

func buscarBloquesContiguosLibres(file *os.File, sb estructuras.SuperBlock, cantidad int, fit byte) ([]int32, error) { // esta funcion busca un tramo contiguo de bloques aplicando el ajuste de la particion
	if cantidad == 0 {
		return nil, nil
	}

	bitmap := make([]byte, sb.BlocksCount)
	if _, err := file.ReadAt(bitmap, int64(sb.BmBlockStart)); err != nil {
		return nil, fmt.Errorf("no se pudo leer bitmap de bloques: %w", err)
	}

	type tramo struct {
		inicio int
		size   int
	}

	var tramos []tramo
	index := 0
	for index < len(bitmap) {
		for index < len(bitmap) && bitmap[index] == '1' {
			index++
		}

		inicio := index
		for index < len(bitmap) && (bitmap[index] == 0 || bitmap[index] == '0') {
			index++
		}

		if index-inicio >= cantidad {
			tramos = append(tramos, tramo{inicio: inicio, size: index - inicio})
		}
	}

	if len(tramos) == 0 {
		return nil, fmt.Errorf("no hay %d bloques contiguos libres", cantidad)
	}

	seleccionado := 0
	for index := 1; index < len(tramos); index++ {
		if fit == 'B' && tramos[index].size < tramos[seleccionado].size {
			seleccionado = index
		}

		if fit == 'W' && tramos[index].size > tramos[seleccionado].size {
			seleccionado = index
		}
	}

	resultado := make([]int32, 0, cantidad)
	for offset := 0; offset < cantidad; offset++ {
		resultado = append(resultado, int32(tramos[seleccionado].inicio+offset))
	}

	return resultado, nil
}

func marcarBloque(file *os.File, sb *estructuras.SuperBlock, index int32, ocupado bool) error { // esta funcion marca un bloque como libre u ocupado
	value := byte('0')
	if ocupado {
		value = '1'
		sb.FreeBlocksCount--
	} else {
		sb.FreeBlocksCount++
	}

	if err := utils.WriteBytesAt(file, int64(sb.BmBlockStart+index), []byte{value}); err != nil {
		return err
	}

	sb.FirstBlock = calcularPrimerBloqueLibre(file, *sb)
	return nil
}

func calcularPrimerBloqueLibre(file *os.File, sb estructuras.SuperBlock) int32 { // esta funcion recalcula el primer bloque libre
	bitmap := make([]byte, sb.BlocksCount)
	if _, err := file.ReadAt(bitmap, int64(sb.BmBlockStart)); err != nil {
		return sb.FirstBlock
	}

	for index, value := range bitmap {
		if value == 0 || value == '0' {
			return int32(index)
		}
	}

	return -1
}
