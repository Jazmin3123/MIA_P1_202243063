package estructuras

type Partition struct { // esta estructura representa una particion guardada dentro del mbr
	Status      byte     // aqui guardo si la particion esta activa o montada
	Type        byte     // aqui guardo el tipo de particion: p, e o l
	Fit         byte     // aqui guardo el ajuste de la particion: b, f o w
	Start       int32    // aqui guardo el byte donde inicia la particion
	Size        int32    // aqui guardo el tamano total de la particion en bytes
	Name        [16]byte // aqui guardo el nombre de la particion con tamano fijo
	Correlative int32    // aqui guardo el correlativo asignado al montar
	Id          [4]byte  // aqui guardo el id de montaje cuando corresponda
}

type MBR struct { // esta estructura representa el master boot record escrito al inicio del disco
	Size          int32        // aqui guardo el tamano total del disco en bytes
	CreationDate  int64        // aqui guardo la fecha de creacion como unix timestamp
	DiskSignature int32        // aqui guardo un numero unico para identificar el disco
	Fit           byte         // aqui guardo el ajuste general del disco
	Partitions    [4]Partition // aqui guardo las cuatro particiones posibles del mbr
}

type EBR struct { // esta estructura representa una particion logica dentro de una extendida
	Mount byte     // aqui guardo si la particion logica esta montada
	Fit   byte     // aqui guardo el ajuste de la particion logica
	Start int32    // aqui guardo el byte donde inicia este ebr o su espacio logico
	Size  int32    // aqui guardo el tamano de la particion logica
	Next  int32    // aqui guardo el byte del siguiente ebr, o -1 si no existe
	Name  [16]byte // aqui guardo el nombre de la particion logica
}
