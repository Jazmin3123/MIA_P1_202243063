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
	size, err := strconv.Atoi(params["size"])
	if err != nil {
		return fmt.Errorf("size debe ser un numero entero")
	}

	_, err = CrearDisco(size, params["unit"], params["fit"], params["path"])
	return err
}

func CrearDisco(size int, unit string, fitValue string, path string) (int64, error) { // esta funcion crea un disco y la reutilizan CLI y API
	sizeBytes, err := parseDiskSize(size, unit) // convierto size y unit a bytes reales
	if err != nil {
		return 0, err
	}

	path = strings.TrimSpace(path) // obtengo la ruta donde se creara el disco
	if path == "" {
		return 0, fmt.Errorf("path es obligatorio")
	}

	if !strings.HasSuffix(strings.ToLower(path), ".mia") {
		return 0, fmt.Errorf("el disco debe tener extension .mia")
	}

	fit, err := parseDiskFit(fitValue) // convierto el ajuste a b, f o w
	if err != nil {
		return 0, err
	}

	if _, err := os.Stat(path); err == nil {
		return 0, fmt.Errorf("ya existe un disco en la ruta %s", path)
	} else if !os.IsNotExist(err) {
		return 0, fmt.Errorf("no se pudo comprobar la ruta del disco: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return 0, fmt.Errorf("no se pudieron crear las carpetas del path: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666) // creo el disco solo si todavia no existe
	if err != nil {
		if os.IsExist(err) {
			return 0, fmt.Errorf("ya existe un disco en la ruta %s", path)
		}
		return 0, fmt.Errorf("no se pudo crear el disco: %w", err)
	}
	defer file.Close() // cierro el archivo al terminar

	if err := fillFileWithZeros(file, sizeBytes); err != nil {
		return 0, err
	}

	mbr := createInitialMBR(sizeBytes, fit) // creo el mbr inicial que va al inicio del disco
	if err := utils.WriteStructAt(file, 0, &mbr); err != nil {
		return 0, err
	}

	fmt.Printf("disco creado correctamente: %s (%d bytes)\n", path, sizeBytes)
	return int64(sizeBytes), nil
}

func parseDiskSize(size int, unit string) (int32, error) { // esta funcion valida size y unit para devolver bytes
	if size <= 0 {
		return 0, fmt.Errorf("size debe ser mayor que cero")
	}

	unit = strings.ToUpper(unit) // si no viene unit, se usa megabytes
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
