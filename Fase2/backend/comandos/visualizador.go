package comandos

import (
	"fmt"
	"os"
	"strings"

	"MIA_P1_202243063/estructuras"
	"MIA_P1_202243063/utils"
)

type DirectoryItem struct { // esta estructura representa una entrada visible para el frontend
	Name  string `json:"name"`
	Type  string `json:"type"`
	Inode int32  `json:"inode"`
}

func ListDirectory(id string, path string) ([]DirectoryItem, error) { // esta funcion lista una carpeta dentro de una particion montada
	mounted, exists := FindMountedPartitionByID(id)
	if !exists {
		return nil, fmt.Errorf("no existe una particion montada con id %s", id)
	}

	path = strings.TrimSpace(path)
	if path == "" {
		path = "/"
	}

	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("path debe ser una ruta absoluta")
	}

	file, err := os.Open(mounted.Path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir el disco: %w", err)
	}
	defer file.Close()

	var sb estructuras.SuperBlock
	if err := utils.ReadStructAt(file, int64(mounted.Partition.Start), &sb); err != nil {
		return nil, err
	}

	if sb.Magic != estructuras.Ext2Magic {
		return nil, fmt.Errorf("la particion no esta formateada como ext2")
	}

	inodeIndex, err := buscarInodoPorRuta(file, sb, path)
	if err != nil {
		return nil, err
	}

	inode, err := leerInodoPorIndice(file, sb, inodeIndex)
	if err != nil {
		return nil, err
	}

	if inode.Type != estructuras.FolderBlockType {
		return nil, fmt.Errorf("%s no es una carpeta", path)
	}

	return listarEntradasCarpeta(file, sb, inode)
}

func ReadFileContent(id string, path string) (string, error) { // esta funcion lee el contenido de un archivo dentro de una particion montada
	mounted, exists := FindMountedPartitionByID(id)
	if !exists {
		return "", fmt.Errorf("no existe una particion montada con id %s", id)
	}

	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("path es obligatorio")
	}

	if !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("path debe ser una ruta absoluta")
	}

	file, err := os.Open(mounted.Path)
	if err != nil {
		return "", fmt.Errorf("no se pudo abrir el disco: %w", err)
	}
	defer file.Close()

	var sb estructuras.SuperBlock
	if err := utils.ReadStructAt(file, int64(mounted.Partition.Start), &sb); err != nil {
		return "", err
	}

	if sb.Magic != estructuras.Ext2Magic {
		return "", fmt.Errorf("la particion no esta formateada como ext2")
	}

	if sesion, activa := obtenerSesionActual(); activa && sesion.Montada.ID == mounted.ID {
		return leerArchivoPorRutaConUsuario(file, sb, path, sesion.Usuario)
	}

	return leerArchivoPorRuta(file, sb, path)
}

func listarEntradasCarpeta(file *os.File, sb estructuras.SuperBlock, inode estructuras.Inode) ([]DirectoryItem, error) {
	items := []DirectoryItem{}

	for blockPointer := 0; blockPointer < 12; blockPointer++ {
		blockIndex := inode.Block[blockPointer]
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

			childInode, err := leerInodoPorIndice(file, sb, content.Inode)
			if err != nil {
				return nil, err
			}

			items = append(items, DirectoryItem{
				Name:  name,
				Type:  directoryItemType(childInode),
				Inode: content.Inode,
			})
		}
	}

	return items, nil
}

func directoryItemType(inode estructuras.Inode) string {
	if inode.Type == estructuras.FolderBlockType {
		return "folder"
	}

	return "file"
}
