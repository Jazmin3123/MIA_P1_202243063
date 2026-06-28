package comandos

import (
	"strings" // uso strings para validar mensajes de error
	"testing" // uso testing para validar el montaje en memoria
)

func TestExecuteMountMountsPrimaryPartition(t *testing.T) { // esta prueba valida que mount genere id para una primaria
	resetMountsForTest()
	path := createTestDisk(t) // creo un disco temporal reutilizando ayuda de fdisk_test

	if err := ExecuteFDisk(map[string]string{"size": "100", "unit": "K", "path": path, "name": "primaria1", "type": "P"}); err != nil {
		t.Fatalf("no se pudo crear primaria: %v", err)
	}

	if err := ExecuteMount(map[string]string{"path": path, "name": "primaria1"}); err != nil {
		t.Fatalf("mount devolvio error: %v", err)
	}

	if len(mountedPartitions) != 1 {
		t.Fatalf("debe existir una particion montada, existen %d", len(mountedPartitions))
	}

	if mountedPartitions[0].ID != "631A" {
		t.Fatalf("el primer id debe ser 631A, fue %s", mountedPartitions[0].ID)
	}
}

func TestExecuteMountRejectsExtendedPartition(t *testing.T) { // esta prueba valida que mount no monte extendidas
	resetMountsForTest()
	path := createTestDisk(t)

	if err := ExecuteFDisk(map[string]string{"size": "300", "unit": "K", "path": path, "name": "extendida1", "type": "E"}); err != nil {
		t.Fatalf("no se pudo crear extendida: %v", err)
	}

	err := ExecuteMount(map[string]string{"path": path, "name": "extendida1"})
	if err == nil {
		t.Fatalf("mount debe rechazar particiones extendidas")
	}

	if !strings.Contains(err.Error(), "no se puede montar una particion extendida: extendida1") {
		t.Fatalf("mensaje inesperado para extendida: %v", err)
	}
}

func TestExecuteMountMountsLogicalPartition(t *testing.T) { // esta prueba valida que mount tambien encuentre particiones logicas
	resetMountsForTest()
	path := createTestDisk(t)

	if err := ExecuteFDisk(map[string]string{"size": "300", "unit": "K", "path": path, "name": "extendida1", "type": "E"}); err != nil {
		t.Fatalf("no se pudo crear extendida: %v", err)
	}

	if err := ExecuteFDisk(map[string]string{"size": "50", "unit": "K", "path": path, "name": "logica1", "type": "L"}); err != nil {
		t.Fatalf("no se pudo crear logica: %v", err)
	}

	if err := ExecuteMount(map[string]string{"path": path, "name": "logica1"}); err != nil {
		t.Fatalf("mount logica devolvio error: %v", err)
	}

	if mountedPartitions[0].Partition.Type != 'L' {
		t.Fatalf("la particion montada debia ser logica, fue %c", mountedPartitions[0].Partition.Type)
	}
}
