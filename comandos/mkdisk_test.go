package comandos

import (
	"os"            // uso os para abrir el disco creado en la prueba
	"path/filepath" // uso filepath para crear rutas temporales
	"testing"       // uso testing para validar el comando mkdisk

	"MIA_P1_202243063/estructuras" // uso estructuras para leer el mbr
	"MIA_P1_202243063/utils"       // uso utils para leer structs binarios
)

func TestExecuteMKDiskCreatesDiskWithMBR(t *testing.T) { // esta prueba valida que mkdisk cree el archivo y escriba mbr
	path := filepath.Join(t.TempDir(), "disco.mia") // creo una ruta temporal para no tocar archivos reales

	err := ExecuteMKDisk(map[string]string{ // ejecuto mkdisk con parametros similares a la cli
		"size": "1",
		"unit": "K",
		"fit":  "BF",
		"path": path,
	})
	if err != nil {
		t.Fatalf("mkdisk devolvio error: %v", err)
	}

	info, err := os.Stat(path) // reviso que el archivo exista
	if err != nil {
		t.Fatalf("no se encontro el disco creado: %v", err)
	}

	if info.Size() != 1024 {
		t.Fatalf("el disco debe medir 1024 bytes, mide %d", info.Size())
	}

	file, err := os.Open(path) // abro el disco para leer el mbr
	if err != nil {
		t.Fatalf("no se pudo abrir el disco: %v", err)
	}
	defer file.Close()

	var mbr estructuras.MBR
	if err := utils.ReadStructAt(file, 0, &mbr); err != nil {
		t.Fatalf("no se pudo leer el mbr: %v", err)
	}

	if mbr.Size != 1024 {
		t.Fatalf("el mbr debe guardar tamano 1024, guardo %d", mbr.Size)
	}

	if mbr.Fit != 'B' {
		t.Fatalf("el mbr debe guardar fit B, guardo %q", mbr.Fit)
	}
}
