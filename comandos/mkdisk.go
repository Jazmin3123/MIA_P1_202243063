package comandos

import (
	"fmt"           // uso fmt para mensajes de error y confirmacion
	"math/rand"     // uso rand para crear la firma unica del disco
	"os"            // uso os para crear carpetas y archivos
	"path/filepath" // uso filepath para obtener la carpeta del path
	"strconv"       // uso strconv para convertir size a numero
	"strings"       // uso strings para validar unit, fit y extension
	"time"          // uso time para fecha de creacion y semilla random

	"MIA_P1_202243063/estructuras" // uso estructuras para crear el mbr
	"MIA_P1_202243063/utils"       // uso utils para escribir structs binarios
)

const zeroBufferSize = 1024 // aqui defino el tamano del bloque de ceros recomendado por el enunciado

func ExecuteMKDisk(params map[string]string) error { // esta funcion ejecuta el comando mkdisk
	sizeBytes, err := parseDiskSize(params) // convierto size y unit a bytes reales
	if err != nil {
		return err
	}

	path := params["path"] // obtengo la ruta donde se creara el disco
	if !strings.HasSuffix(strings.ToLower(path), ".mia") {
		return fmt.Errorf("el disco debe tener extension .mia")
	}

	fit, err := parseDiskFit(params["fit"]) // convierto el ajuste a b, f o w
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("no se pudieron crear las carpetas del path: %w", err)
	}

	file, err := os.Create(path) // creo el archivo del disco desde cero
	if err != nil {
		return fmt.Errorf("no se pudo crear el disco: %w", err)
	}
	defer file.Close() // cierro el archivo al terminar

	if err := fillFileWithZeros(file, sizeBytes); err != nil {
		return err
	}

	mbr := createInitialMBR(sizeBytes, fit) // creo el mbr inicial que va al inicio del disco
	if err := utils.WriteStructAt(file, 0, &mbr); err != nil {
		return err
	}

	fmt.Printf("disco creado correctamente: %s (%d bytes)\n", path, sizeBytes)
	return nil
}

func parseDiskSize(params map[string]string) (int32, error) { // esta funcion valida size y unit para devolver bytes
	size, err := strconv.Atoi(params["size"])
	if err != nil {
		return 0, fmt.Errorf("size debe ser un numero entero")
	}

	if size <= 0 {
		return 0, fmt.Errorf("size debe ser mayor que cero")
	}

	unit := strings.ToUpper(params["unit"]) // si no viene unit, se usa megabytes
	if unit == "" {
		unit = "M"
	}

	switch unit {
	case "K":
		size *= 1024
	case "M":
		size *= 1024 * 1024
	default:
		return 0, fmt.Errorf("unit debe ser K o M")
	}

	return int32(size), nil
}

func parseDiskFit(value string) (byte, error) { // esta funcion valida el ajuste del disco
	fit := strings.ToUpper(value) // si no viene fit, se usa first fit
	if fit == "" {
		fit = "FF"
	}

	switch fit {
	case "BF":
		return 'B', nil
	case "FF":
		return 'F', nil
	case "WF":
		return 'W', nil
	default:
		return 0, fmt.Errorf("fit debe ser BF, FF o WF")
	}
}

func fillFileWithZeros(file *os.File, size int32) error { // esta funcion llena el archivo con ceros binarios
	buffer := make([]byte, zeroBufferSize) // uso un buffer de 1024 bytes para no escribir byte por byte
	remaining := int64(size)

	for remaining > 0 {
		chunkSize := int64(len(buffer))
		if remaining < chunkSize {
			chunkSize = remaining
		}

		if _, err := file.Write(buffer[:chunkSize]); err != nil {
			return fmt.Errorf("no se pudo llenar el disco con ceros: %w", err)
		}

		remaining -= chunkSize
	}

	return nil
}

func createInitialMBR(size int32, fit byte) estructuras.MBR { // esta funcion arma el mbr inicial del disco
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	return estructuras.MBR{
		Size:          size,                       // guardo el tamano total del disco
		CreationDate:  time.Now().Unix(),          // guardo la fecha actual en formato unix
		DiskSignature: random.Int31(),             // guardo una firma random para identificar el disco
		Fit:           fit,                        // guardo el ajuste elegido
		Partitions:    [4]estructuras.Partition{}, // dejo las particiones vacias al inicio
	}
}
