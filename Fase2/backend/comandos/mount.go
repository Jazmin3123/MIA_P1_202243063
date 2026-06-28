package comandos

import (
	"fmt"           // uso fmt para crear mensajes y ids
	"os"            // uso os para abrir el disco
	"path/filepath" // uso filepath para normalizar rutas

	"MIA_P1_202243063/estructuras" // uso estructuras para leer mbr y particiones
	"MIA_P1_202243063/utils"       // uso utils para leer structs y convertir nombres
)

const carnetSuffix = "63" // aqui uso los ultimos dos digitos del carnet 202243063

type MountedPartition struct { // esta estructura representa una particion montada en memoria
	ID        string                // aqui guardo el id generado, por ejemplo 631A
	Path      string                // aqui guardo la ruta del disco
	Name      string                // aqui guardo el nombre de la particion
	DiskIndex int                   // aqui guardo el numero asignado al disco
	Letter    byte                  // aqui guardo la letra asignada dentro del disco
	Partition estructuras.Partition // aqui guardo la informacion de la particion montada
}

type MountedPartitionInfo struct { // esta estructura expone montajes de forma amigable para la API
	ID          string `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Letter      string `json:"letter"`
	Correlative int    `json:"correlative"`
	DiskIndex   int    `json:"disk_index"`
	Type        string `json:"type"`
	Start       int32  `json:"start"`
	Size        int32  `json:"size"`
}

var mountedPartitions []MountedPartition  // aqui guardo las particiones montadas mientras el programa esta abierto
var mountedDiskNumbers = map[string]int{} // aqui guardo el numero asignado a cada disco montado
var nextDiskNumber = 1                    // aqui guardo el siguiente numero disponible para un disco nuevo

func ExecuteMount(params map[string]string) error { // esta funcion ejecuta el comando mount
	path, err := filepath.Abs(params["path"]) // convierto la ruta a absoluta para evitar duplicados por rutas distintas
	if err != nil {
		return fmt.Errorf("no se pudo resolver la ruta del disco: %w", err)
	}

	file, err := os.Open(path) // abro el disco en solo lectura porque mount trabaja en memoria
	if err != nil {
		return fmt.Errorf("no se pudo abrir el disco: %w", err)
	}
	defer file.Close()

	var mbr estructuras.MBR
	if err := utils.ReadStructAt(file, 0, &mbr); err != nil {
		return err
	}

	name := params["name"] // obtengo el nombre de la particion a montar
	partition, err := findMountablePartitionByName(file, mbr, name)
	if err != nil {
		return err
	}

	if mounted := findMountedPartition(path, name); mounted != nil {
		fmt.Printf("particion ya montada con id %s\n", mounted.ID)
		printMountedPartitions()
		return nil
	}

	diskIndex := diskNumberForPath(path)                         // obtengo o creo el numero del disco
	letter := nextMountLetterForDisk(path)                       // obtengo la siguiente letra disponible para este disco
	id := fmt.Sprintf("%s%d%c", carnetSuffix, diskIndex, letter) // construyo el id con carnet, disco y letra

	mountedPartitions = append(mountedPartitions, MountedPartition{
		ID:        id,        // guardo el id generado
		Path:      path,      // guardo la ruta normalizada
		Name:      name,      // guardo el nombre de la particion
		DiskIndex: diskIndex, // guardo el numero de disco
		Letter:    letter,    // guardo la letra asignada
		Partition: partition, // guardo la particion encontrada
	})

	fmt.Printf("particion montada correctamente: %s id=%s\n", name, id)
	printMountedPartitions()
	return nil
}

func findMountablePartitionByName(file *os.File, mbr estructuras.MBR, name string) (estructuras.Partition, error) { // esta funcion busca una particion primaria o logica por nombre
	for _, partition := range mbr.Partitions {
		if partition.Size == 0 {
			continue
		}

		if utils.BytesToString(partition.Name[:]) != name {
			continue
		}

		if partition.Type == 'E' {
			return estructuras.Partition{}, fmt.Errorf("no se puede montar una particion extendida: %s", name)
		}

		if partition.Type == 'P' {
			return partition, nil
		}
	}

	logical, found, err := findLogicalPartitionByName(file, mbr, name)
	if err != nil {
		return estructuras.Partition{}, err
	}

	if found {
		return logical, nil
	}

	return estructuras.Partition{}, fmt.Errorf("no existe una particion primaria o logica con nombre %s", name)
}

func findLogicalPartitionByName(file *os.File, mbr estructuras.MBR, name string) (estructuras.Partition, bool, error) { // esta funcion recorre los ebr para encontrar una particion logica
	extended, found := findExtendedPartition(mbr)
	if !found {
		return estructuras.Partition{}, false, nil
	}

	ebrSize := utils.StructSize(estructuras.EBR{})
	currentPosition := extended.Start
	for currentPosition != -1 {
		var ebr estructuras.EBR
		if err := utils.ReadStructAt(file, int64(currentPosition), &ebr); err != nil {
			return estructuras.Partition{}, false, err
		}

		if ebr.Size > 0 && utils.BytesToString(ebr.Name[:]) == name {
			partition := estructuras.Partition{
				Status:      ebr.Mount,                // copio el estado guardado en el ebr
				Type:        'L',                      // identifico que viene de una particion logica
				Fit:         ebr.Fit,                  // copio el ajuste de la logica
				Start:       ebr.Start + ebrSize,      // el espacio util empieza despues del ebr
				Size:        ebr.Size,                 // guardo el tamano util de la logica
				Name:        ebr.Name,                 // copio el nombre fijo
				Correlative: 0,                        // el correlativo se maneja en memoria al montar
				Id:          utils.StringToBytes4(""), // el id real se guarda en la tabla de montajes
			}
			return partition, true, nil
		}

		currentPosition = ebr.Next
	}

	return estructuras.Partition{}, false, nil
}

func findMountedPartition(path string, name string) *MountedPartition { // esta funcion revisa si la particion ya esta montada
	for index := range mountedPartitions {
		if mountedPartitions[index].Path == path && mountedPartitions[index].Name == name {
			return &mountedPartitions[index]
		}
	}

	return nil
}

func diskNumberForPath(path string) int { // esta funcion devuelve el numero de disco usado para el id
	if number, exists := mountedDiskNumbers[path]; exists {
		return number
	}

	mountedDiskNumbers[path] = nextDiskNumber
	nextDiskNumber++
	return mountedDiskNumbers[path]
}

func nextMountLetterForDisk(path string) byte { // esta funcion calcula la siguiente letra disponible para el disco
	count := 0
	for _, mounted := range mountedPartitions {
		if mounted.Path == path {
			count++
		}
	}

	return byte('A' + count)
}

func printMountedPartitions() { // esta funcion muestra la tabla simple de particiones montadas
	fmt.Println("particiones montadas:")
	for _, mounted := range mountedPartitions {
		fmt.Printf("  id=%s path=%s name=%s\n", mounted.ID, mounted.Path, mounted.Name)
	}
}

func FindMountedPartitionByID(id string) (*MountedPartition, bool) { // esta funcion busca una particion montada usando su id
	for index := range mountedPartitions {
		if mountedPartitions[index].ID == id {
			return &mountedPartitions[index], true
		}
	}

	return nil, false
}

func ListMountedPartitions() interface{} { // esta funcion devuelve los montajes actuales para consultas de la API
	mounts := make([]MountedPartitionInfo, 0, len(mountedPartitions))

	for _, mounted := range mountedPartitions {
		mounts = append(mounts, MountedPartitionInfo{
			ID:          mounted.ID,
			Name:        mounted.Name,
			Path:        mounted.Path,
			Letter:      string(mounted.Letter),
			Correlative: mounted.DiskIndex,
			DiskIndex:   mounted.DiskIndex,
			Type:        string(mounted.Partition.Type),
			Start:       mounted.Partition.Start,
			Size:        mounted.Partition.Size,
		})
	}

	return mounts
}

func resetMountsForTest() { // esta funcion limpia montajes para pruebas automaticas
	mountedPartitions = nil
	mountedDiskNumbers = map[string]int{}
	nextDiskNumber = 1
}
