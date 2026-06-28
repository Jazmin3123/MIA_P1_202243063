package comandos

import (
	"fmt"     // uso fmt para mensajes de error y confirmacion
	"os"      // uso os para abrir el disco binario
	"sort"    // uso sort para ordenar particiones por inicio
	"strconv" // uso strconv para convertir size a numero
	"strings" // uso strings para validar type, fit y unit

	"MIA_P1_202243063/estructuras" // uso estructuras para leer y escribir mbr y ebr
	"MIA_P1_202243063/utils"       // uso utils para conversiones y lectura binaria
)

type partitionSpace struct { // esta estructura me ayuda a calcular espacios libres dentro del disco
	Start int32 // aqui guardo donde inicia el espacio libre
	Size  int32 // aqui guardo cuantos bytes libres hay
}

func ExecuteFDisk(params map[string]string) error { // esta funcion ejecuta el comando fdisk para crear particiones
	sizeBytes, err := parseFDiskSize(params) // convierto size y unit a bytes reales
	if err != nil {
		return err
	}

	partType, err := parsePartitionType(params["type"]) // valido si sera primaria, extendida o logica
	if err != nil {
		return err
	}

	partFit, err := parsePartitionFit(params["fit"]) // valido el ajuste de la particion
	if err != nil {
		return err
	}

	name := params["name"] // obtengo el nombre de la particion
	if len(name) > 16 {
		return fmt.Errorf("name no debe superar 16 caracteres")
	}

	file, err := os.OpenFile(params["path"], os.O_RDWR, 0666) // abro el disco para modificar el mbr
	if err != nil {
		return fmt.Errorf("no se pudo abrir el disco: %w", err)
	}
	defer file.Close()

	var mbr estructuras.MBR
	if err := utils.ReadStructAt(file, 0, &mbr); err != nil {
		return err
	}

	if err := ensureUniquePartitionName(file, mbr, name); err != nil {
		return err
	}

	if partType == 'L' {
		return createLogicalPartition(file, mbr, sizeBytes, partFit, name)
	}

	return createPrimaryOrExtendedPartition(file, &mbr, sizeBytes, partType, partFit, name)
}

func parseFDiskSize(params map[string]string) (int32, error) { // esta funcion valida size y unit para fdisk
	size, err := strconv.Atoi(params["size"])
	if err != nil {
		return 0, fmt.Errorf("size debe ser un numero entero")
	}

	if size <= 0 {
		return 0, fmt.Errorf("size debe ser mayor que cero")
	}

	unit := strings.ToUpper(params["unit"]) // si no viene unit, se usa kilobytes
	if unit == "" {
		unit = "K"
	}

	switch unit {
	case "B":
	case "K":
		size *= 1024
	case "M":
		size *= 1024 * 1024
	default:
		return 0, fmt.Errorf("unit debe ser B, K o M")
	}

	return int32(size), nil
}

func parsePartitionType(value string) (byte, error) { // esta funcion valida el tipo de particion
	partType := strings.ToUpper(value) // si no viene type, se usa primaria
	if partType == "" {
		partType = "P"
	}

	switch partType {
	case "P":
		return 'P', nil
	case "E":
		return 'E', nil
	case "L":
		return 'L', nil
	default:
		return 0, fmt.Errorf("type debe ser P, E o L")
	}
}

func parsePartitionFit(value string) (byte, error) { // esta funcion valida el ajuste de una particion
	fit := strings.ToUpper(value) // si no viene fit, se usa worst fit
	if fit == "" {
		fit = "WF"
	}

	switch fit {
	case "BF", "BESTFIT":
		return 'B', nil
	case "FF", "FIRSTFIT":
		return 'F', nil
	case "WF", "WORSTFIT":
		return 'W', nil
	default:
		return 0, fmt.Errorf("fit debe ser BF, FF, WF, BestFit, FirstFit o WorstFit")
	}
}

func createPrimaryOrExtendedPartition(file *os.File, mbr *estructuras.MBR, size int32, partType byte, fit byte, name string) error { // esta funcion crea particiones primarias o extendidas
	slot := firstEmptyPartitionSlot(*mbr) // busco una posicion libre en las cuatro particiones del mbr
	if slot == -1 {
		return fmt.Errorf("ya existen cuatro particiones primarias/extendidas")
	}

	if partType == 'E' && hasExtendedPartition(*mbr) {
		return fmt.Errorf("ya existe una particion extendida en el disco")
	}

	start, err := findPartitionStart(*mbr, size, fit) // busco espacio libre aplicando el ajuste elegido
	if err != nil {
		return err
	}

	mbr.Partitions[slot] = estructuras.Partition{
		Status:      0,                           // la particion inicia sin montar
		Type:        partType,                    // guardo si es primaria o extendida
		Fit:         fit,                         // guardo el ajuste de esta particion
		Start:       start,                       // guardo el byte de inicio
		Size:        size,                        // guardo el tamano en bytes
		Name:        utils.StringToBytes16(name), // guardo el nombre como arreglo fijo
		Correlative: 0,                           // todavia no tiene correlativo de montaje
		Id:          [4]byte{},                   // todavia no tiene id de montaje
	}

	if err := utils.WriteStructAt(file, 0, mbr); err != nil {
		return err
	}

	if partType == 'E' {
		emptyEBR := estructuras.EBR{Next: -1, Start: start, Fit: fit} // creo el primer ebr vacio dentro de la extendida
		if err := utils.WriteStructAt(file, int64(start), &emptyEBR); err != nil {
			return err
		}
	}

	fmt.Printf("particion %s creada correctamente en %d (%d bytes)\n", name, start, size)
	return nil
}

func firstEmptyPartitionSlot(mbr estructuras.MBR) int { // esta funcion busca el primer espacio vacio en el arreglo de particiones
	for index, partition := range mbr.Partitions {
		if partition.Size == 0 {
			return index
		}
	}

	return -1
}

func hasExtendedPartition(mbr estructuras.MBR) bool { // esta funcion revisa si el disco ya tiene particion extendida
	for _, partition := range mbr.Partitions {
		if partition.Size > 0 && partition.Type == 'E' {
			return true
		}
	}

	return false
}

func findPartitionStart(mbr estructuras.MBR, size int32, fit byte) (int32, error) { // esta funcion calcula donde colocar la nueva particion
	spaces := freePrimarySpaces(mbr) // obtengo los espacios libres fuera de particiones existentes
	if len(spaces) == 0 {
		return 0, fmt.Errorf("no hay espacio libre en el disco")
	}

	selected := -1
	for index, space := range spaces {
		if space.Size < size {
			continue
		}

		if selected == -1 {
			selected = index
			continue
		}

		if fit == 'B' && space.Size < spaces[selected].Size {
			selected = index
		}

		if fit == 'W' && space.Size > spaces[selected].Size {
			selected = index
		}
	}

	if selected == -1 {
		return 0, fmt.Errorf("no hay espacio suficiente para la particion")
	}

	return spaces[selected].Start, nil
}

func freePrimarySpaces(mbr estructuras.MBR) []partitionSpace { // esta funcion obtiene huecos libres entre el mbr y las particiones
	used := usedPrimaryPartitions(mbr) // obtengo particiones existentes ordenadas por inicio
	mbrSize := utils.StructSize(estructuras.MBR{})
	currentStart := mbrSize
	var spaces []partitionSpace

	for _, partition := range used {
		if partition.Start > currentStart {
			spaces = append(spaces, partitionSpace{Start: currentStart, Size: partition.Start - currentStart})
		}

		currentStart = partition.Start + partition.Size
	}

	if currentStart < mbr.Size {
		spaces = append(spaces, partitionSpace{Start: currentStart, Size: mbr.Size - currentStart})
	}

	return spaces
}

func usedPrimaryPartitions(mbr estructuras.MBR) []estructuras.Partition { // esta funcion devuelve las particiones ocupadas del mbr
	var used []estructuras.Partition

	for _, partition := range mbr.Partitions {
		if partition.Size > 0 {
			used = append(used, partition)
		}
	}

	sort.Slice(used, func(i, j int) bool { // ordeno por inicio para calcular espacios libres
		return used[i].Start < used[j].Start
	})

	return used
}

func ensureUniquePartitionName(file *os.File, mbr estructuras.MBR, name string) error { // esta funcion evita nombres repetidos en el disco
	for _, partition := range mbr.Partitions {
		if partition.Size == 0 {
			continue
		}

		if utils.BytesToString(partition.Name[:]) == name {
			return fmt.Errorf("ya existe una particion con el nombre %s", name)
		}
	}

	return ensureUniqueLogicalName(file, mbr, name)
}

func createLogicalPartition(file *os.File, mbr estructuras.MBR, size int32, fit byte, name string) error { // esta funcion crea una particion logica dentro de la extendida
	extended, found := findExtendedPartition(mbr)
	if !found {
		return fmt.Errorf("no existe una particion extendida para crear logicas")
	}

	ebrSize := utils.StructSize(estructuras.EBR{})
	var first estructuras.EBR
	if err := utils.ReadStructAt(file, int64(extended.Start), &first); err != nil {
		return err
	}

	if first.Size == 0 {
		if ebrSize+size > extended.Size {
			return fmt.Errorf("no hay espacio suficiente dentro de la extendida")
		}

		first = estructuras.EBR{
			Mount: 0,                           // la logica inicia sin montar
			Fit:   fit,                         // guardo el ajuste de la logica
			Start: extended.Start,              // guardo la posicion del ebr
			Size:  size,                        // guardo el tamano util de la logica
			Next:  -1,                          // no hay siguiente ebr
			Name:  utils.StringToBytes16(name), // guardo el nombre como arreglo fijo
		}

		if err := utils.WriteStructAt(file, int64(extended.Start), &first); err != nil {
			return err
		}

		fmt.Printf("particion logica %s creada correctamente en %d (%d bytes)\n", name, first.Start, size)
		return nil
	}

	currentPosition := extended.Start
	current := first

	for current.Next != -1 {
		currentPosition = current.Next
		if err := utils.ReadStructAt(file, int64(currentPosition), &current); err != nil {
			return err
		}
	}

	newStart := current.Start + ebrSize + current.Size
	if newStart+ebrSize+size > extended.Start+extended.Size {
		return fmt.Errorf("no hay espacio suficiente dentro de la extendida")
	}

	current.Next = newStart
	if err := utils.WriteStructAt(file, int64(currentPosition), &current); err != nil {
		return err
	}

	newEBR := estructuras.EBR{
		Mount: 0,                           // la logica inicia sin montar
		Fit:   fit,                         // guardo el ajuste de la logica
		Start: newStart,                    // guardo donde queda el nuevo ebr
		Size:  size,                        // guardo el tamano util de la logica
		Next:  -1,                          // queda como ultimo nodo de la lista
		Name:  utils.StringToBytes16(name), // guardo el nombre como arreglo fijo
	}

	if err := utils.WriteStructAt(file, int64(newStart), &newEBR); err != nil {
		return err
	}

	fmt.Printf("particion logica %s creada correctamente en %d (%d bytes)\n", name, newStart, size)
	return nil
}

func findExtendedPartition(mbr estructuras.MBR) (estructuras.Partition, bool) { // esta funcion devuelve la particion extendida si existe
	for _, partition := range mbr.Partitions {
		if partition.Size > 0 && partition.Type == 'E' {
			return partition, true
		}
	}

	return estructuras.Partition{}, false
}

func ensureUniqueLogicalName(file *os.File, mbr estructuras.MBR, name string) error { // esta funcion revisa nombres repetidos en particiones logicas
	extended, found := findExtendedPartition(mbr)
	if !found {
		return nil
	}

	currentPosition := extended.Start
	for currentPosition != -1 {
		var ebr estructuras.EBR
		if err := utils.ReadStructAt(file, int64(currentPosition), &ebr); err != nil {
			return err
		}

		if ebr.Size > 0 && utils.BytesToString(ebr.Name[:]) == name {
			return fmt.Errorf("ya existe una particion con el nombre %s", name)
		}

		currentPosition = ebr.Next
	}

	return nil
}
