package estructuras

import (
	"encoding/binary" // uso binary para calcular el tamano real de los structs
	"testing"         // uso testing para validar las estructuras del proyecto
)

func TestBlockSizes(t *testing.T) { // esta prueba confirma que los bloques ext2 pesan 64 bytes
	if size := binary.Size(FolderBlock{}); size != 64 {
		t.Fatalf("folderblock debe medir 64 bytes, mide %d", size)
	}

	if size := binary.Size(FileBlock{}); size != 64 {
		t.Fatalf("fileblock debe medir 64 bytes, mide %d", size)
	}

	if size := binary.Size(PointerBlock{}); size != 64 {
		t.Fatalf("pointerblock debe medir 64 bytes, mide %d", size)
	}
}
