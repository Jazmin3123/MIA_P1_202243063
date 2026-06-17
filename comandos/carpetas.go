package comandos

import (
	"fmt"     // uso fmt para errores y confirmaciones
	"os"      // uso os para abrir el disco ext2
	"strings" // uso strings para limpiar y partir rutas
	"time"    // uso time para actualizar fechas de inodos

	"MIA_P1_202243063/estructuras" // uso estructuras para inodos y bloques de carpeta
	"MIA_P1_202243063/utils"       // uso utils para leer/escribir binario y convertir nombres
)

func EjecutarMKDIR(params map[string]string, flags map[string]bool) error { // esta funcion crea carpetas dentro del sistema ext2
	sesion, activa := obtenerSesionActual()
	if !activa {
		return fmt.Errorf("no existe una sesion activa")
	}

	path := params["path"] // obtengo la ruta absoluta de la carpeta a crear
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path debe ser una ruta absoluta")
	}

	partes := separarRuta(path)
	if len(partes) == 0 {
		return fmt.Errorf("no se puede crear la carpeta raiz")
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

	if sb.Magic != estructuras.Ext2Magic {
		return fmt.Errorf("la particion no esta formateada como ext2")
	}

	parentIndex := int32(0) // empiezo desde la raiz
	for index, nombre := range partes {
		if len(nombre) > 12 {
			return fmt.Errorf("cada carpeta debe tener maximo 12 caracteres: %s", nombre)
		}

		childIndex, exists, err := buscarEntradaEnCarpeta(file, sb, parentIndex, nombre)
		if err != nil {
			return err
		}

		if exists {
			parentIndex = childIndex
			continue
		}

		esUltima := index == len(partes)-1
		if !esUltima && !flags["p"] {
			return fmt.Errorf("no existe la carpeta padre %s, usa -p para crear padres", nombre)
		}

		newIndex, err := crearCarpeta(file, &sb, parentIndex, nombre, sesion.Usuario.UID, sesion.Usuario.GID)
		if err != nil {
			return err
		}

		parentIndex = newIndex
	}

	if err := utils.WriteStructAt(file, int64(sesion.Montada.Partition.Start), &sb); err != nil {
		return err
	}

	fmt.Printf("carpeta creada correctamente: %s\n", path)
	return nil
}

func separarRuta(path string) []string { // esta funcion separa una ruta absoluta en nombres utiles
	partesCrudas := strings.Split(path, "/")
	var partes []string

	for _, parte := range partesCrudas {
		parte = strings.TrimSpace(parte)
		if parte != "" {
			partes = append(partes, parte)
		}
	}

	return partes
}

func buscarEntradaEnCarpeta(file *os.File, sb estructuras.SuperBlock, inodeIndex int32, nombre string) (int32, bool, error) { // esta funcion busca un nombre dentro de una carpeta
	inode, err := leerInodoPorIndice(file, sb, inodeIndex)
	if err != nil {
		return 0, false, err
	}

	if inode.Type != estructuras.FolderBlockType {
		return 0, false, fmt.Errorf("la ruta contiene un archivo donde se esperaba carpeta")
	}

	for blockPointer := 0; blockPointer < 12; blockPointer++ {
		blockIndex := inode.Block[blockPointer]
		if blockIndex == -1 {
			continue
		}

		block, err := leerBloqueCarpeta(file, sb, blockIndex)
		if err != nil {
			return 0, false, err
		}

		for _, content := range block.Content {
			if content.Inode != -1 && utils.BytesToString(content.Name[:]) == nombre {
				return content.Inode, true, nil
			}
		}
	}

	return 0, false, nil
}

func crearCarpeta(file *os.File, sb *estructuras.SuperBlock, parentIndex int32, nombre string, uid int32, gid int32) (int32, error) { // esta funcion reserva inodo y bloque para una carpeta nueva
	inodeIndex, err := buscarInodoLibre(file, *sb)
	if err != nil {
		return 0, err
	}

	blockIndex, err := buscarBloqueLibre(file, *sb)
	if err != nil {
		return 0, err
	}

	if err := marcarInodo(file, sb, inodeIndex, true); err != nil {
		return 0, err
	}

	if err := marcarBloque(file, sb, blockIndex, true); err != nil {
		return 0, err
	}

	inode := createInitialInode(uid, gid, 0, estructuras.FolderBlockType, "664")
	inode.Block[0] = blockIndex

	block := estructuras.FolderBlock{
		Content: [4]estructuras.Content{
			{Name: utils.StringToBytes12("."), Inode: inodeIndex},
			{Name: utils.StringToBytes12(".."), Inode: parentIndex},
			{Name: [12]byte{}, Inode: -1},
			{Name: [12]byte{}, Inode: -1},
		},
	}

	if err := escribirEntradaEnCarpeta(file, sb, parentIndex, nombre, inodeIndex); err != nil {
		return 0, err
	}

	if err := escribirInodoPorIndice(file, *sb, inodeIndex, inode); err != nil {
		return 0, err
	}

	if err := escribirBloqueCarpeta(file, *sb, blockIndex, block); err != nil {
		return 0, err
	}

	return inodeIndex, nil
}

func escribirEntradaEnCarpeta(file *os.File, sb *estructuras.SuperBlock, parentIndex int32, nombre string, childIndex int32) error { // esta funcion agrega una entrada al bloque carpeta padre
	parent, err := leerInodoPorIndice(file, *sb, parentIndex)
	if err != nil {
		return err
	}

	for pointerIndex := 0; pointerIndex < 12; pointerIndex++ {
		blockIndex := parent.Block[pointerIndex]
		if blockIndex == -1 {
			newBlock, err := buscarBloqueLibre(file, *sb)
			if err != nil {
				return err
			}

			if err := marcarBloque(file, sb, newBlock, true); err != nil {
				return err
			}

			parent.Block[pointerIndex] = newBlock
			empty := bloqueCarpetaVacio()
			empty.Content[0] = estructuras.Content{Name: utils.StringToBytes12(nombre), Inode: childIndex}

			if err := escribirBloqueCarpeta(file, *sb, newBlock, empty); err != nil {
				return err
			}

			parent.MTime = time.Now().Unix()
			return escribirInodoPorIndice(file, *sb, parentIndex, parent)
		}

		block, err := leerBloqueCarpeta(file, *sb, blockIndex)
		if err != nil {
			return err
		}

		for contentIndex := range block.Content {
			if block.Content[contentIndex].Inode == -1 {
				block.Content[contentIndex] = estructuras.Content{Name: utils.StringToBytes12(nombre), Inode: childIndex}
				if err := escribirBloqueCarpeta(file, *sb, blockIndex, block); err != nil {
					return err
				}

				parent.MTime = time.Now().Unix()
				return escribirInodoPorIndice(file, *sb, parentIndex, parent)
			}
		}
	}

	return fmt.Errorf("la carpeta padre no tiene espacio para mas entradas directas")
}

func bloqueCarpetaVacio() estructuras.FolderBlock { // esta funcion crea un bloque carpeta con entradas libres
	return estructuras.FolderBlock{
		Content: [4]estructuras.Content{
			{Name: [12]byte{}, Inode: -1},
			{Name: [12]byte{}, Inode: -1},
			{Name: [12]byte{}, Inode: -1},
			{Name: [12]byte{}, Inode: -1},
		},
	}
}

func leerInodoPorIndice(file *os.File, sb estructuras.SuperBlock, index int32) (estructuras.Inode, error) { // esta funcion lee un inodo por indice
	var inode estructuras.Inode
	position := int64(sb.InodeStart + index*sb.InodeSize)
	if err := utils.ReadStructAt(file, position, &inode); err != nil {
		return estructuras.Inode{}, err
	}

	return inode, nil
}

func escribirInodoPorIndice(file *os.File, sb estructuras.SuperBlock, index int32, inode estructuras.Inode) error { // esta funcion escribe un inodo por indice
	position := int64(sb.InodeStart + index*sb.InodeSize)
	return utils.WriteStructAt(file, position, &inode)
}

func leerBloqueCarpeta(file *os.File, sb estructuras.SuperBlock, index int32) (estructuras.FolderBlock, error) { // esta funcion lee un bloque carpeta por indice
	var block estructuras.FolderBlock
	position := int64(sb.BlockStart + index*sb.BlockSize)
	if err := utils.ReadStructAt(file, position, &block); err != nil {
		return estructuras.FolderBlock{}, err
	}

	return block, nil
}

func escribirBloqueCarpeta(file *os.File, sb estructuras.SuperBlock, index int32, block estructuras.FolderBlock) error { // esta funcion escribe un bloque carpeta por indice
	position := int64(sb.BlockStart + index*sb.BlockSize)
	return utils.WriteStructAt(file, position, &block)
}

func buscarInodoLibre(file *os.File, sb estructuras.SuperBlock) (int32, error) { // esta funcion busca el primer inodo libre en el bitmap
	bitmap := make([]byte, sb.InodesCount)
	if _, err := file.ReadAt(bitmap, int64(sb.BmInodeStart)); err != nil {
		return 0, fmt.Errorf("no se pudo leer bitmap de inodos: %w", err)
	}

	for index, value := range bitmap {
		if value == 0 || value == '0' {
			return int32(index), nil
		}
	}

	return 0, fmt.Errorf("no hay inodos libres")
}

func marcarInodo(file *os.File, sb *estructuras.SuperBlock, index int32, ocupado bool) error { // esta funcion marca un inodo como libre u ocupado
	value := byte('0')
	if ocupado {
		value = '1'
		sb.FreeInodesCount--
	} else {
		sb.FreeInodesCount++
	}

	if err := utils.WriteBytesAt(file, int64(sb.BmInodeStart+index), []byte{value}); err != nil {
		return err
	}

	sb.FirstInode = calcularPrimerInodoLibre(file, *sb)
	return nil
}

func calcularPrimerInodoLibre(file *os.File, sb estructuras.SuperBlock) int32 { // esta funcion recalcula el primer inodo libre
	bitmap := make([]byte, sb.InodesCount)
	if _, err := file.ReadAt(bitmap, int64(sb.BmInodeStart)); err != nil {
		return sb.FirstInode
	}

	for index, value := range bitmap {
		if value == 0 || value == '0' {
			return int32(index)
		}
	}

	return -1
}
