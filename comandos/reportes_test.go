package comandos

import (
	"os"            // uso os para leer reportes generados
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
}

func leerReportePrueba(t *testing.T, path string) string { // esta funcion lee un reporte generado en pruebas
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("no se pudo leer reporte %s: %v", path, err)
	}

	return string(content)
}
