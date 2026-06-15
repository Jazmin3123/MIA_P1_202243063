package estructuras

const (
	Ext2Magic       int32 = 0xEF53 // aqui guardo el valor magico que identifica ext2
	FolderBlockType byte  = 0      // aqui represento un inodo de carpeta
	FileBlockType   byte  = 1      // aqui represento un inodo de archivo
)

type SuperBlock struct { // esta estructura guarda la informacion principal del sistema ext2
	FilesystemType  int32 // aqui guardo el tipo de sistema de archivos
	InodesCount     int32 // aqui guardo la cantidad total de inodos
	BlocksCount     int32 // aqui guardo la cantidad total de bloques
	FreeBlocksCount int32 // aqui guardo la cantidad de bloques libres
	FreeInodesCount int32 // aqui guardo la cantidad de inodos libres
	MTime           int64 // aqui guardo la ultima fecha de montaje como unix timestamp
	UMTime          int64 // aqui guardo la ultima fecha de desmontaje como unix timestamp
	MntCount        int32 // aqui guardo cuantas veces se ha montado el sistema
	Magic           int32 // aqui guardo el valor 0xef53 para identificar ext2
	InodeSize       int32 // aqui guardo el tamano de la estructura inode
	BlockSize       int32 // aqui guardo el tamano de los bloques
	FirstInode      int32 // aqui guardo el primer inodo libre
	FirstBlock      int32 // aqui guardo el primer bloque libre
	BmInodeStart    int32 // aqui guardo donde inicia el bitmap de inodos
	BmBlockStart    int32 // aqui guardo donde inicia el bitmap de bloques
	InodeStart      int32 // aqui guardo donde inicia la tabla de inodos
	BlockStart      int32 // aqui guardo donde inicia la tabla de bloques
}

type Inode struct { // esta estructura guarda la informacion de un archivo o carpeta
	UID   int32     // aqui guardo el id del usuario propietario
	GID   int32     // aqui guardo el id del grupo propietario
	Size  int32     // aqui guardo el tamano del archivo en bytes
	ATime int64     // aqui guardo la ultima fecha de lectura como unix timestamp
	CTime int64     // aqui guardo la fecha de creacion como unix timestamp
	MTime int64     // aqui guardo la ultima fecha de modificacion como unix timestamp
	Block [15]int32 // aqui guardo apuntadores a bloques directos e indirectos
	Type  byte      // aqui guardo si el inodo es carpeta o archivo
	Perm  [3]byte   // aqui guardo permisos ugo en forma octal, por ejemplo 664
}

type Content struct { // esta estructura representa una entrada dentro de un bloque de carpeta
	Name  [12]byte // aqui guardo el nombre del archivo o carpeta
	Inode int32    // aqui guardo el apuntador al inodo asociado
}

type FolderBlock struct { // esta estructura representa un bloque de carpeta de 64 bytes
	Content [4]Content // aqui guardo hasta cuatro entradas de carpeta
}

type FileBlock struct { // esta estructura representa un bloque de archivo de 64 bytes
	Content [64]byte // aqui guardo el contenido del archivo
}

type PointerBlock struct { // esta estructura representa un bloque de apuntadores indirectos
	Pointers [16]int32 // aqui guardo apuntadores hacia otros bloques
}
