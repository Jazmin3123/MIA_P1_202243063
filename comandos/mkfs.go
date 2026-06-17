package comandos

import (
	"fmt"     // uso fmt para mensajes de error y confirmacion
	"os"      // uso os para abrir el disco a formatear
	"strings" // uso strings para validar el tipo de formateo
	"time"    // uso time para fechas del superbloque e inodos

	"MIA_P1_202243063/estructuras" // uso estructuras para escribir ext2
	"MIA_P1_202243063/utils"       // uso utils para escribir structs binarios
)

const initialUsersContent = "1,G,root\n1,U,root,root,123\n" // aqui dejo el contenido inicial de users.txt

func ExecuteMKFS(params map[string]string) error { // esta funcion ejecuta el comando mkfs
	formatType := strings.ToLower(params["type"]) // si no viene type, se usa full
	if formatType == "" {
		formatType = "full"
	}

	if formatType != "full" {
		return fmt.Errorf("mkfs solo admite type=full")
	}

	mounted, exists := FindMountedPartitionByID(params["id"]) // busco la particion montada por id
	if !exists {
		return fmt.Errorf("no existe una particion montada con id %s", params["id"])
	}

	file, err := os.OpenFile(mounted.Path, os.O_RDWR, 0666) // abro el disco para escribir ext2
	if err != nil {
		return fmt.Errorf("no se pudo abrir el disco: %w", err)
	}
	defer file.Close()

	layout, err := buildEXT2Layout(mounted.Partition) // calculo posiciones y cantidades del sistema ext2
	if err != nil {
		return err
	}

	if err := writeEXT2InitialState(file, layout); err != nil {
		return err
	}

	fmt.Printf("particion %s formateada como ext2 con id %s\n", mounted.Name, mounted.ID)
	return nil
}

type ext2Layout struct { // esta estructura me ayuda a mover los datos calculados para mkfs
	SuperBlock estructuras.SuperBlock // aqui guardo el superbloque ya calculado
	InodeCount int32                  // aqui guardo cuantos inodos se reservaron
	BlockCount int32                  // aqui guardo cuantos bloques se reservaron
}

func buildEXT2Layout(partition estructuras.Partition) (ext2Layout, error) { // esta funcion calcula la distribucion ext2 dentro de la particion
	superBlockSize := utils.StructSize(estructuras.SuperBlock{}) // tamano del superbloque
	inodeSize := utils.StructSize(estructuras.Inode{})           // tamano de cada inodo
	blockSize := utils.StructSize(estructuras.FileBlock{})       // tamano de cada bloque, debe ser 64
	denominator := int32(4) + inodeSize + int32(3)*blockSize     // aqui aplico n + 3n + n*inode + 3n*block
	inodeCount := (partition.Size - superBlockSize) / denominator

	if inodeCount < 2 {
		return ext2Layout{}, fmt.Errorf("la particion es muy pequena para formatear ext2")
	}

	blockCount := int32(3) * inodeCount
	now := time.Now().Unix()
	bmInodeStart := partition.Start + superBlockSize
	bmBlockStart := bmInodeStart + inodeCount
	inodeStart := bmBlockStart + blockCount
	blockStart := inodeStart + inodeCount*inodeSize

	superBlock := estructuras.SuperBlock{
		FilesystemType:  2,                     // identifico el sistema como ext2
		InodesCount:     inodeCount,            // guardo el total de inodos
		BlocksCount:     blockCount,            // guardo el total de bloques
		FreeBlocksCount: blockCount - 2,        // uso dos bloques iniciales: raiz y users.txt
		FreeInodesCount: inodeCount - 2,        // uso dos inodos iniciales: raiz y users.txt
		MTime:           now,                   // fecha de montaje/formateo
		UMTime:          0,                     // aun no hay desmontaje
		MntCount:        1,                     // primer montaje/formateo registrado
		Magic:           estructuras.Ext2Magic, // valor magico ext2
		InodeSize:       inodeSize,             // tamano de inodo
		BlockSize:       blockSize,             // tamano de bloque
		FirstInode:      2,                     // el primer inodo libre queda despues de raiz y users
		FirstBlock:      2,                     // el primer bloque libre queda despues de raiz y users
		BmInodeStart:    bmInodeStart,          // inicio bitmap inodos
		BmBlockStart:    bmBlockStart,          // inicio bitmap bloques
		InodeStart:      inodeStart,            // inicio tabla inodos
		BlockStart:      blockStart,            // inicio tabla bloques
	}

	return ext2Layout{SuperBlock: superBlock, InodeCount: inodeCount, BlockCount: blockCount}, nil
}

func writeEXT2InitialState(file *os.File, layout ext2Layout) error { // esta funcion escribe las estructuras iniciales de ext2
	sb := layout.SuperBlock

	if err := utils.WriteStructAt(file, int64(sb.BmInodeStart), make([]byte, layout.InodeCount)); err != nil {
		return err
	}

	if err := utils.WriteStructAt(file, int64(sb.BmBlockStart), make([]byte, layout.BlockCount)); err != nil {
		return err
	}

	if err := utils.WriteBytesAt(file, int64(sb.BmInodeStart), []byte{'1', '1'}); err != nil {
		return err
	}

	if err := utils.WriteBytesAt(file, int64(sb.BmBlockStart), []byte{'1', '1'}); err != nil {
		return err
	}

	if err := utils.WriteStructAt(file, int64(sb.BmBlockStart+layout.BlockCount), make([]byte, layout.InodeCount*sb.InodeSize)); err != nil {
		return err
	}

	if err := utils.WriteStructAt(file, int64(sb.BlockStart), make([]byte, layout.BlockCount*sb.BlockSize)); err != nil {
		return err
	}

	rootInode := createInitialInode(1, 1, 0, estructuras.FolderBlockType, "777") // creo el inodo de la carpeta raiz
	rootInode.Block[0] = 0

	usersInode := createInitialInode(1, 1, int32(len(initialUsersContent)), estructuras.FileBlockType, "664") // creo el inodo users.txt
	usersInode.Block[0] = 1

	rootBlock := createRootFolderBlock()        // creo el bloque de carpeta raiz
	usersBlock := createInitialUsersFileBlock() // creo el bloque de archivo users.txt

	if err := utils.WriteStructAt(file, int64(sb.InodeStart), &rootInode); err != nil {
		return err
	}

	if err := utils.WriteStructAt(file, int64(sb.InodeStart+sb.InodeSize), &usersInode); err != nil {
		return err
	}

	if err := utils.WriteStructAt(file, int64(sb.BlockStart), &rootBlock); err != nil {
		return err
	}

	if err := utils.WriteStructAt(file, int64(sb.BlockStart+sb.BlockSize), &usersBlock); err != nil {
		return err
	}

	if err := utils.WriteStructAt(file, int64(sb.BmInodeStart-(utils.StructSize(estructuras.SuperBlock{}))), &sb); err != nil {
		return err
	}

	return nil
}

func createInitialInode(uid int32, gid int32, size int32, inodeType byte, perm string) estructuras.Inode { // esta funcion crea un inodo inicial con bloques en -1
	now := time.Now().Unix()
	inode := estructuras.Inode{
		UID:   uid,                        // usuario propietario
		GID:   gid,                        // grupo propietario
		Size:  size,                       // tamano del contenido
		ATime: now,                        // fecha de acceso
		CTime: now,                        // fecha de creacion
		MTime: now,                        // fecha de modificacion
		Type:  inodeType,                  // tipo archivo o carpeta
		Perm:  utils.StringToBytes3(perm), // permisos ugo
	}

	for index := range inode.Block {
		inode.Block[index] = -1 // inicializo todos los apuntadores como no usados
	}

	return inode
}

func createRootFolderBlock() estructuras.FolderBlock { // esta funcion crea el bloque inicial de la carpeta raiz
	return estructuras.FolderBlock{
		Content: [4]estructuras.Content{
			{Name: utils.StringToBytes12("."), Inode: 0},         // entrada para la carpeta actual
			{Name: utils.StringToBytes12(".."), Inode: 0},        // entrada para la carpeta padre
			{Name: utils.StringToBytes12("users.txt"), Inode: 1}, // entrada del archivo de usuarios
			{Name: [12]byte{}, Inode: -1},                        // entrada libre
		},
	}
}

func createInitialUsersFileBlock() estructuras.FileBlock { // esta funcion crea el bloque inicial de users.txt
	var block estructuras.FileBlock
	copy(block.Content[:], []byte(initialUsersContent))
	return block
}
