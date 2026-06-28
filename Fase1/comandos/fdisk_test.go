package comandos

import (
	"os"            // uso os para abrir el disco creado en la prueba
	"path/filepath" // uso filepath para crear rutas temporales
	"testing"       // uso testing para validar fdisk

	"MIA_P1_202243063/estructuras" // uso estructuras para leer mbr y ebr
	"MIA_P1_202243063/utils"       // uso utils para leer structs binarios
)

func TestExecuteFDiskCreatesPrimaryPartition(t *testing.T) { // esta prueba valida que fdisk escriba una particion primaria en el mbr
	path := createTestDisk(t) // creo un disco temporal para la prueba

	err := ExecuteFDisk(map[string]string{ // ejecuto fdisk con una particion primaria
		"size": "100",
		"unit": "K",
		"path": path,
		"name": "primaria1",
		"type": "P",
	})
	if err != nil {
		t.Fatalf("fdisk devolvio error: %v", err)
	}

	mbr := readTestMBR(t, path) // leo el mbr para validar la particion escrita
	partition := mbr.Partitions[0]

	if partition.Size != 100*1024 {
		t.Fatalf("la particion debe medir 102400 bytes, mide %d", partition.Size)
	}

	if partition.Type != 'P' {
		t.Fatalf("la particion debe ser primaria, guardo %q", partition.Type)
	}

	if utils.BytesToString(partition.Name[:]) != "primaria1" {
		t.Fatalf("la particion debe llamarse primaria1, guardo %q", utils.BytesToString(partition.Name[:]))
	}
}

func TestExecuteFDiskCreatesExtendedAndLogicalPartition(t *testing.T) { // esta prueba valida extendida y logica con ebr
	path := createTestDisk(t) // creo un disco temporal para la prueba

	if err := ExecuteFDisk(map[string]string{"size": "300", "unit": "K", "path": path, "name": "extendida1", "type": "E"}); err != nil {
		t.Fatalf("no se pudo crear extendida: %v", err)
	}

	if err := ExecuteFDisk(map[string]string{"size": "50", "unit": "K", "path": path, "name": "logica1", "type": "L"}); err != nil {
		t.Fatalf("no se pudo crear logica: %v", err)
	}

	mbr := readTestMBR(t, path) // leo el mbr para encontrar la extendida
	extended := mbr.Partitions[0]

	file, err := os.Open(path) // abro el disco para leer el primer ebr
	if err != nil {
		t.Fatalf("no se pudo abrir el disco: %v", err)
	}
	defer file.Close()

	var ebr estructuras.EBR
	if err := utils.ReadStructAt(file, int64(extended.Start), &ebr); err != nil {
		t.Fatalf("no se pudo leer el ebr: %v", err)
	}

	if ebr.Size != 50*1024 {
		t.Fatalf("la logica debe medir 51200 bytes, mide %d", ebr.Size)
	}

	if utils.BytesToString(ebr.Name[:]) != "logica1" {
		t.Fatalf("la logica debe llamarse logica1, guardo %q", utils.BytesToString(ebr.Name[:]))
	}
}

func createTestDisk(t *testing.T) string { // esta funcion crea un disco temporal para pruebas de particiones
	t.Helper()
	path := filepath.Join(t.TempDir(), "disco.mia")

	if err := ExecuteMKDisk(map[string]string{"size": "1", "unit": "M", "path": path}); err != nil {
		t.Fatalf("no se pudo crear disco de prueba: %v", err)
	}

	return path
}

func readTestMBR(t *testing.T, path string) estructuras.MBR { // esta funcion lee el mbr de un disco de prueba
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("no se pudo abrir el disco: %v", err)
	}
	defer file.Close()

	var mbr estructuras.MBR
	if err := utils.ReadStructAt(file, 0, &mbr); err != nil {
		t.Fatalf("no se pudo leer el mbr: %v", err)
	}

	return mbr
}
