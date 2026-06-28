package comandos

import (
	"os"            // uso os para leer reportes generados
	"os/exec"       // uso exec para validar sintaxis dot cuando graphviz existe
	"path/filepath" // uso filepath para rutas temporales
	"strings"       // uso strings para validar contenido
	"testing"       // uso testing para validar reportes
)

func TestEjecutarREPGeneraReportesBasicos(t *testing.T) { // esta prueba valida reportes sb, bitmap y file
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKFILE(map[string]string{"path": "/docs/a.txt", "size": "5"}, map[string]bool{"r": true}); err != nil {
		t.Fatalf("mkfile devolvio error: %v", err)
	}

	dir := t.TempDir()
	if err := EjecutarREP(map[string]string{"id": "631A", "name": "sb", "path": filepath.Join(dir, "sb.dot")}); err != nil {
		t.Fatalf("rep sb devolvio error: %v", err)
	}

	if err := EjecutarREP(map[string]string{"id": "631A", "name": "bm_inode", "path": filepath.Join(dir, "bm_inode.txt")}); err != nil {
		t.Fatalf("rep bm_inode devolvio error: %v", err)
	}

	if err := EjecutarREP(map[string]string{"id": "631A", "name": "file", "path": filepath.Join(dir, "file.txt"), "path_file_ls": "/docs/a.txt"}); err != nil {
		t.Fatalf("rep file devolvio error: %v", err)
	}

	if err := EjecutarREP(map[string]string{"id": "631A", "name": "inode", "path": filepath.Join(dir, "inode.dot")}); err != nil {
		t.Fatalf("rep inode devolvio error: %v", err)
	}

	if err := EjecutarREP(map[string]string{"id": "631A", "name": "block", "path": filepath.Join(dir, "block.dot")}); err != nil {
		t.Fatalf("rep block devolvio error: %v", err)
	}

	if err := EjecutarREP(map[string]string{"id": "631A", "name": "ls", "path": filepath.Join(dir, "ls.dot"), "path_file_ls": "/docs"}); err != nil {
		t.Fatalf("rep ls devolvio error: %v", err)
	}

	if err := EjecutarREP(map[string]string{"id": "631A", "name": "tree", "path": filepath.Join(dir, "tree.dot")}); err != nil {
		t.Fatalf("rep tree devolvio error: %v", err)
	}

	sbContent := leerReportePrueba(t, filepath.Join(dir, "sb.dot"))
	if !strings.Contains(sbContent, "SUPER BLOQUE") {
		t.Fatalf("reporte sb no contiene titulo esperado: %q", sbContent)
	}

	bitmapContent := leerReportePrueba(t, filepath.Join(dir, "bm_inode.txt"))
	if !strings.Contains(bitmapContent, "1 1") {
		t.Fatalf("bitmap no contiene inodos iniciales ocupados: %q", bitmapContent)
	}

	fileContent := leerReportePrueba(t, filepath.Join(dir, "file.txt"))
	if !strings.Contains(fileContent, "01234") {
		t.Fatalf("reporte file no contiene datos esperados: %q", fileContent)
	}

	inodeContent := leerReportePrueba(t, filepath.Join(dir, "inode.dot"))
	if !strings.Contains(inodeContent, "INODO") {
		t.Fatalf("reporte inode no contiene inodos: %q", inodeContent)
	}

	blockContent := leerReportePrueba(t, filepath.Join(dir, "block.dot"))
	if !strings.Contains(blockContent, "BLOQUE") {
		t.Fatalf("reporte block no contiene bloques: %q", blockContent)
	}

	lsContent := leerReportePrueba(t, filepath.Join(dir, "ls.dot"))
	if !strings.Contains(lsContent, "a.txt") {
		t.Fatalf("reporte ls no contiene archivo esperado: %q", lsContent)
	}

	treeContent := leerReportePrueba(t, filepath.Join(dir, "tree.dot"))
	if !strings.Contains(treeContent, "inode0") || !strings.Contains(treeContent, "block") {
		t.Fatalf("reporte tree no contiene arbol esperado: %q", treeContent)
	}
}

func TestEjecutarREPIncluyeParticionesLogicas(t *testing.T) { // esta prueba valida que disk y mbr muestren ebr y logicas
	resetMountsForTest()
	limpiarSesionParaPrueba()
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

	dir := t.TempDir()
	if err := EjecutarREP(map[string]string{"id": "631A", "name": "disk", "path": filepath.Join(dir, "disk.dot")}); err != nil {
		t.Fatalf("rep disk devolvio error: %v", err)
	}

	if err := EjecutarREP(map[string]string{"id": "631A", "name": "mbr", "path": filepath.Join(dir, "mbr.dot")}); err != nil {
		t.Fatalf("rep mbr devolvio error: %v", err)
	}

	diskContent := leerReportePrueba(t, filepath.Join(dir, "disk.dot"))
	if !strings.Contains(diskContent, "Logica") || !strings.Contains(diskContent, "logica1") {
		t.Fatalf("reporte disk no contiene logica esperada: %q", diskContent)
	}

	if _, err := exec.LookPath("dot"); err == nil {
		cmd := exec.Command("dot", "-Tsvg", filepath.Join(dir, "disk.dot"), "-o", filepath.Join(dir, "disk.svg"))
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("graphviz rechazo reporte disk: %v: %s", err, output)
		}
	}

	mbrContent := leerReportePrueba(t, filepath.Join(dir, "mbr.dot"))
	if !strings.Contains(mbrContent, "EBR") || !strings.Contains(mbrContent, "logica1") {
		t.Fatalf("reporte mbr no contiene ebr esperado: %q", mbrContent)
	}
}

func leerReportePrueba(t *testing.T, path string) string { // esta funcion lee un reporte generado en pruebas
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("no se pudo leer reporte %s: %v", path, err)
	}

	return string(content)
}
