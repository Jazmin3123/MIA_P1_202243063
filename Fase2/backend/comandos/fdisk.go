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

func ExecuteFDisk(params map[string]string) error { // esta funcion ejecuta el comando fdisk para crear, eliminar o ajustar particiones
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

	if deleteMode := params["delete"]; deleteMode != "" {
		return deletePartition(file, &mbr, name, deleteMode)
	}

	if addValue := params["add"]; addValue != "" {
		return resizePartition(file, &mbr, name, addValue, params["unit"])
	}

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

func parseFDiskDelta(value string, unitValue string) (int32, error) { // esta funcion valida -add con signo y unidad
	if strings.TrimSpace(value) == "" {
		return 0, fmt.Errorf("add debe ser un numero entero diferente de cero")
	}

	size, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("add debe ser un numero entero")
	}

	if size == 0 {
		return 0, fmt.Errorf("add debe ser diferente de cero")
	}

	unit := strings.ToUpper(unitValue)
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

func deletePartition(file *os.File, mbr *estructuras.MBR, name string, mode string) error { // esta funcion elimina particiones primarias, extendidas o logicas
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode != "fast" && mode != "full" {
		return fmt.Errorf("delete debe ser fast o full")
	}

	if index := findPrimaryPartitionIndex(*mbr, name); index != -1 {
		partition := mbr.Partitions[index]
		if mode == "full" {
			if err := zeroDiskRange(file, partition.Start, partition.Size); err != nil {
				return err
			}
		}

		mbr.Partitions[index] = estructuras.Partition{}
		if err := utils.WriteStructAt(file, 0, mbr); err != nil {
			return err
		}

		if partition.Type == 'E' {
			fmt.Printf("particion extendida %s eliminada correctamente con delete=%s; sus logicas quedaron eliminadas\n", name, mode)
			return nil
		}

		fmt.Printf("particion %s eliminada correctamente con delete=%s\n", name, mode)
		return nil
	}

	logicalPosition, logical, found, err := findLogicalPartitionByExactName(file, *mbr, name)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("no existe una particion con nombre %s", name)
	}

	if mode == "full" {
		ebrSize := utils.StructSize(estructuras.EBR{})
		if err := zeroDiskRange(file, logical.Start+ebrSize, logical.Size); err != nil {
			return err
		}
	}

	logical.Mount = 0
	logical.Fit = 0
	logical.Size = 0
	logical.Name = [16]byte{}
	if err := utils.WriteStructAt(file, int64(logicalPosition), &logical); err != nil {
		return err
	}

	fmt.Printf("particion logica %s eliminada correctamente con delete=%s\n", name, mode)
	return nil
}

func resizePartition(file *os.File, mbr *estructuras.MBR, name string, addValue string, unit string) error { // esta funcion aumenta o reduce una particion existente
	delta, err := parseFDiskDelta(addValue, unit)
	if err != nil {
		return err
	}

	if index := findPrimaryPartitionIndex(*mbr, name); index != -1 {
		return resizePrimaryPartition(file, mbr, index, delta)
	}

	logicalPosition, logical, found, err := findLogicalPartitionByExactName(file, *mbr, name)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("no existe una particion con nombre %s", name)
	}

	return resizeLogicalPartition(file, *mbr, logicalPosition, logical, delta)
}

func resizePrimaryPartition(file *os.File, mbr *estructuras.MBR, index int, delta int32) error { // esta funcion ajusta una particion del mbr
	partition := mbr.Partitions[index]
	newSize := partition.Size + delta
	if newSize <= 0 {
		return fmt.Errorf("add reduce demasiado la particion %s: el tamano resultante debe ser mayor que cero", utils.BytesToString(partition.Name[:]))
	}

	if delta > 0 {
		freeAfter := freeSpaceAfterPrimary(*mbr, index)
		if freeAfter < delta {
			return fmt.Errorf("no hay espacio libre suficiente despues de la particion %s: disponible %d bytes, requerido %d bytes", utils.BytesToString(partition.Name[:]), freeAfter, delta)
		}
	}

	if partition.Type == 'E' && delta < 0 {
		if err := validateExtendedShrink(file, partition, newSize); err != nil {
			return err
		}
	}

	mbr.Partitions[index].Size = newSize
	if err := utils.WriteStructAt(file, 0, mbr); err != nil {
		return err
	}

	fmt.Printf("particion %s ajustada correctamente: nuevo tamano %d bytes\n", utils.BytesToString(partition.Name[:]), newSize)
	return nil
}

func resizeLogicalPartition(file *os.File, mbr estructuras.MBR, position int32, logical estructuras.EBR, delta int32) error { // esta funcion ajusta una particion logica
	newSize := logical.Size + delta
	name := utils.BytesToString(logical.Name[:])
	if newSize <= 0 {
		return fmt.Errorf("add reduce demasiado la particion logica %s: el tamano resultante debe ser mayor que cero", name)
	}

	if delta > 0 {
		freeAfter, err := freeSpaceAfterLogical(mbr, logical)
		if err != nil {
			return err
		}
		if freeAfter < delta {
			return fmt.Errorf("no hay espacio libre suficiente despues de la particion logica %s: disponible %d bytes, requerido %d bytes", name, freeAfter, delta)
		}
	}

	logical.Size = newSize
	if err := utils.WriteStructAt(file, int64(position), &logical); err != nil {
		return err
	}

	fmt.Printf("particion logica %s ajustada correctamente: nuevo tamano %d bytes\n", name, newSize)
	return nil
}

func findPrimaryPartitionIndex(mbr estructuras.MBR, name string) int { // esta funcion busca primarias o extendidas por nombre
	for index, partition := range mbr.Partitions {
		if partition.Size > 0 && utils.BytesToString(partition.Name[:]) == name {
			return index
		}
	}

	return -1
}

func findLogicalPartitionByExactName(file *os.File, mbr estructuras.MBR, name string) (int32, estructuras.EBR, bool, error) { // esta funcion busca una logica y devuelve su ebr
	extended, found := findExtendedPartition(mbr)
	if !found {
		return 0, estructuras.EBR{}, false, nil
	}

	currentPosition := extended.Start
	for currentPosition != -1 {
		var ebr estructuras.EBR
		if err := utils.ReadStructAt(file, int64(currentPosition), &ebr); err != nil {
			return 0, estructuras.EBR{}, false, err
		}

		if ebr.Size > 0 && utils.BytesToString(ebr.Name[:]) == name {
			return currentPosition, ebr, true, nil
		}

		currentPosition = ebr.Next
	}

	return 0, estructuras.EBR{}, false, nil
}

func freeSpaceAfterPrimary(mbr estructuras.MBR, index int) int32 { // esta funcion calcula espacio libre contiguo despues de una particion del mbr
	partition := mbr.Partitions[index]
	end := partition.Start + partition.Size
	limit := mbr.Size

	for otherIndex, other := range mbr.Partitions {
		if otherIndex == index || other.Size == 0 {
			continue
		}

		if other.Start >= end && other.Start < limit {
			limit = other.Start
		}
	}

	return limit - end
}

func freeSpaceAfterLogical(mbr estructuras.MBR, logical estructuras.EBR) (int32, error) { // esta funcion calcula espacio libre contiguo despues de una logica
	extended, found := findExtendedPartition(mbr)
	if !found {
		return 0, fmt.Errorf("no existe una particion extendida")
	}

	ebrSize := utils.StructSize(estructuras.EBR{})
	end := logical.Start + ebrSize + logical.Size
	if logical.Next != -1 {
		return logical.Next - end, nil
	}

	return extended.Start + extended.Size - end, nil
}

func validateExtendedShrink(file *os.File, extended estructuras.Partition, newSize int32) error { // esta funcion evita cortar logicas al reducir una extendida
	newEnd := extended.Start + newSize
	currentPosition := extended.Start
	ebrSize := utils.StructSize(estructuras.EBR{})

	for currentPosition != -1 {
		var ebr estructuras.EBR
		if err := utils.ReadStructAt(file, int64(currentPosition), &ebr); err != nil {
			return err
		}

		if ebr.Size > 0 && ebr.Start+ebrSize+ebr.Size > newEnd {
			return fmt.Errorf("no se puede reducir la extendida: la logica %s quedaria fuera del nuevo tamano", utils.BytesToString(ebr.Name[:]))
		}

		currentPosition = ebr.Next
	}

	return nil
}

func zeroDiskRange(file *os.File, start int32, size int32) error { // esta funcion rellena un rango del disco con ceros
	if size <= 0 {
		return nil
	}

	buffer := make([]byte, zeroBufferSize)
	remaining := int64(size)
	position := int64(start)

	for remaining > 0 {
		chunkSize := int64(len(buffer))
		if remaining < chunkSize {
			chunkSize = remaining
		}

		if _, err := file.WriteAt(buffer[:chunkSize], position); err != nil {
			return fmt.Errorf("no se pudo limpiar el espacio de la particion: %w", err)
		}

		position += chunkSize
		remaining -= chunkSize
	}

	return nil
}
