package comandos

import (
	"bufio"         // uso bufio para simular respuestas de consola
	"io"            // uso io para leer la salida capturada de cat
	"os"            // uso os para abrir el disco y crear archivo cont temporal
	"path/filepath" // uso filepath para rutas temporales
	"strings"       // uso strings para preparar entrada simulada
	"testing"       // uso testing para validar mkfile y cat

	"MIA_P1_202243063/estructuras" // uso estructuras para leer superbloque
	"MIA_P1_202243063/utils"       // uso utils para leer structs binarios
)

func TestEjecutarMKFILECreaArchivoConSize(t *testing.T) { // esta prueba valida mkfile con contenido generado por size
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKFILE(map[string]string{"path": "/home/user/a.txt", "size": "15"}, map[string]bool{"r": true}); err != nil {
		t.Fatalf("mkfile devolvio error: %v", err)
	}

	contenido := leerArchivoPrueba(t, "/home/user/a.txt")
	if contenido != "012345678901234" {
		t.Fatalf("contenido inesperado: %q", contenido)
	}
}

func TestEjecutarMKFILEConContTienePrioridad(t *testing.T) { // esta prueba valida que -cont tenga prioridad sobre -size
	prepararSesionRootParaUsuarios(t)
	externo := filepath.Join(t.TempDir(), "entrada.txt")
	if err := os.WriteFile(externo, []byte("contenido externo"), 0644); err != nil {
		t.Fatalf("no se pudo crear archivo externo: %v", err)
	}

	if err := EjecutarMKFILE(map[string]string{"path": "/b.txt", "size": "99", "cont": externo}, map[string]bool{}); err != nil {
		t.Fatalf("mkfile devolvio error: %v", err)
	}

	contenido := leerArchivoPrueba(t, "/b.txt")
	if contenido != "contenido externo" {
		t.Fatalf("contenido inesperado: %q", contenido)
	}
}

func TestLeerArchivoPorRutaParaCAT(t *testing.T) { // esta prueba valida la lectura que usa cat
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKFILE(map[string]string{"path": "/docs/c.txt", "size": "5"}, map[string]bool{"r": true}); err != nil {
		t.Fatalf("mkfile devolvio error: %v", err)
	}

	contenido := leerArchivoPrueba(t, "/docs/c.txt")
	if contenido != "01234" {
		t.Fatalf("cat/lectura esperaba 01234, obtuvo %q", contenido)
	}
}

func TestEjecutarMKFILENoSobrescribeSinConfirmacion(t *testing.T) { // esta prueba valida que mkfile pregunte antes de sobrescribir
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKFILE(map[string]string{"path": "/a.txt", "size": "5"}, map[string]bool{}); err != nil {
		t.Fatalf("mkfile inicial devolvio error: %v", err)
	}

	oldReader := consoleReader
	consoleReader = bufio.NewReader(strings.NewReader("n\n"))
	defer func() { consoleReader = oldReader }()

	if err := EjecutarMKFILE(map[string]string{"path": "/a.txt", "size": "10"}, map[string]bool{}); err != nil {
		t.Fatalf("mkfile repetido devolvio error: %v", err)
	}

	contenido := leerArchivoPrueba(t, "/a.txt")
	if contenido != "01234" {
		t.Fatalf("el archivo no debio sobrescribirse, contenido=%q", contenido)
	}
}

func TestEjecutarCATContinuaSiUnArchivoNoExiste(t *testing.T) { // esta prueba valida que cat separe el error y no interrumpa la lista
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKFILE(map[string]string{"path": "/a.txt", "size": "5"}, map[string]bool{}); err != nil {
		t.Fatalf("mkfile devolvio error: %v", err)
	}

	salida := capturarSalidaPrueba(t, func() {
		if err := EjecutarCAT(map[string]string{"file1": "/a.txt", "file2": "/no_existe.txt"}); err != nil {
			t.Fatalf("cat no debio detenerse: %v", err)
		}
	})

	if !strings.Contains(salida, "01234\nError: no existe la ruta /no_existe.txt") {
		t.Fatalf("salida de cat inesperada: %q", salida)
	}
}

func capturarSalidaPrueba(t *testing.T, accion func()) string { // esta funcion captura stdout para validar mensajes de comandos
	t.Helper()
	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("no se pudo crear pipe: %v", err)
	}
	os.Stdout = writer

	accion()

	writer.Close()
	os.Stdout = oldStdout
	salida, err := io.ReadAll(reader)
	reader.Close()
	if err != nil {
		t.Fatalf("no se pudo leer salida: %v", err)
	}

	return string(salida)
}

func leerArchivoPrueba(t *testing.T, path string) string { // esta funcion lee un archivo del ext2 de prueba
	t.Helper()
	file, err := os.Open(sesionActual.Montada.Path)
	if err != nil {
		t.Fatalf("no se pudo abrir disco: %v", err)
	}
	defer file.Close()

	var sb estructuras.SuperBlock
	if err := utils.ReadStructAt(file, int64(sesionActual.Montada.Partition.Start), &sb); err != nil {
		t.Fatalf("no se pudo leer superbloque: %v", err)
	}

	contenido, err := leerArchivoPorRuta(file, sb, path)
	if err != nil {
		t.Fatalf("no se pudo leer archivo %s: %v", path, err)
	}

	return contenido
}
