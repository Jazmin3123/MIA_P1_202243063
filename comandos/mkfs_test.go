package comandos

import (
	"os"      // uso os para abrir el disco formateado
	"strings" // uso strings para validar el contenido inicial
	"testing" // uso testing para validar mkfs

	"MIA_P1_202243063/estructuras" // uso estructuras para leer superbloque y bloques
	"MIA_P1_202243063/utils"       // uso utils para leer structs binarios
)

func TestExecuteMKFSFormatsMountedPartition(t *testing.T) { // esta prueba valida el formateo ext2 inicial
	resetMountsForTest()
	path := createTestDisk(t) // creo un disco temporal

	if err := ExecuteFDisk(map[string]string{"size": "500", "unit": "K", "path": path, "name": "primaria1", "type": "P"}); err != nil {
		t.Fatalf("no se pudo crear primaria: %v", err)
	}

	if err := ExecuteMount(map[string]string{"path": path, "name": "primaria1"}); err != nil {
		t.Fatalf("no se pudo montar primaria: %v", err)
	}

	if err := ExecuteMKFS(map[string]string{"id": "631A"}); err != nil {
		t.Fatalf("mkfs devolvio error: %v", err)
	}

	mbr := readTestMBR(t, path)
	partition := mbr.Partitions[0]

	file, err := os.Open(path) // abro el disco para leer estructuras ext2
	if err != nil {
		t.Fatalf("no se pudo abrir el disco: %v", err)
	}
	defer file.Close()

	var sb estructuras.SuperBlock
	if err := utils.ReadStructAt(file, int64(partition.Start), &sb); err != nil {
		t.Fatalf("no se pudo leer el superbloque: %v", err)
	}

	if sb.Magic != estructuras.Ext2Magic {
		t.Fatalf("magic esperado %x, obtenido %x", estructuras.Ext2Magic, sb.Magic)
	}

	if sb.InodesCount <= 0 || sb.BlocksCount != sb.InodesCount*3 {
		t.Fatalf("conteos invalidos: inodos=%d bloques=%d", sb.InodesCount, sb.BlocksCount)
	}

	bitmap := make([]byte, 2)
	if _, err := file.ReadAt(bitmap, int64(sb.BmInodeStart)); err != nil {
		t.Fatalf("no se pudo leer bitmap de inodos: %v", err)
	}

	if string(bitmap) != "11" {
		t.Fatalf("los primeros dos inodos deben estar ocupados, bitmap=%q", string(bitmap))
	}

	var usersBlock estructuras.FileBlock
	if err := utils.ReadStructAt(file, int64(sb.BlockStart+sb.BlockSize), &usersBlock); err != nil {
		t.Fatalf("no se pudo leer bloque users.txt: %v", err)
	}

	content := utils.BytesToString(usersBlock.Content[:])
	if !strings.Contains(content, "1,G,root") || !strings.Contains(content, "1,U,root,root,123") {
		t.Fatalf("users.txt inicial no contiene usuarios esperados: %q", content)
	}
}
