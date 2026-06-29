package comandos

import (
	"os"
	"path/filepath"
	"testing"

	"MIA_P1_202243063/estructuras"
	"MIA_P1_202243063/utils"
)

func TestEjecutarRENAMERenombraArchivo(t *testing.T) { // esta prueba valida rename sobre un archivo existente
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKFILE(map[string]string{"path": "/docs/a.txt", "size": "4"}, map[string]bool{"r": true}); err != nil {
		t.Fatalf("mkfile devolvio error: %v", err)
	}

	if err := EjecutarRENAME(map[string]string{"path": "/docs/a.txt", "name": "b.txt"}); err != nil {
		t.Fatalf("rename devolvio error: %v", err)
	}

	if contenido := leerArchivoPrueba(t, "/docs/b.txt"); contenido != "0123" {
		t.Fatalf("contenido renombrado inesperado: %q", contenido)
	}

	if _, err := leerArchivoPruebaConError(t, "/docs/a.txt"); err == nil {
		t.Fatalf("la ruta anterior no debe existir despues de rename")
	}
}

func TestEjecutarEDITReemplazaContenido(t *testing.T) { // esta prueba valida edit leyendo contenido desde el sistema operativo
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKFILE(map[string]string{"path": "/docs/a.txt", "size": "4"}, map[string]bool{"r": true}); err != nil {
		t.Fatalf("mkfile devolvio error: %v", err)
	}

	externo := filepath.Join(t.TempDir(), "contenido.txt")
	if err := os.WriteFile(externo, []byte("nuevo contenido"), 0644); err != nil {
		t.Fatalf("no se pudo escribir contenido externo: %v", err)
	}

	if err := EjecutarEDIT(map[string]string{"path": "/docs/a.txt", "contenido": externo}); err != nil {
		t.Fatalf("edit devolvio error: %v", err)
	}

	if contenido := leerArchivoPrueba(t, "/docs/a.txt"); contenido != "nuevo contenido" {
		t.Fatalf("contenido editado inesperado: %q", contenido)
	}
}

func TestEjecutarREMOVEEliminaCarpetaRecursiva(t *testing.T) { // esta prueba valida remove recursivo sobre carpetas
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKFILE(map[string]string{"path": "/docs/sub/a.txt", "size": "5"}, map[string]bool{"r": true}); err != nil {
		t.Fatalf("mkfile devolvio error: %v", err)
	}

	if err := EjecutarREMOVE(map[string]string{"path": "/docs"}); err != nil {
		t.Fatalf("remove devolvio error: %v", err)
	}

	if _, err := leerArchivoPruebaConError(t, "/docs/sub/a.txt"); err == nil {
		t.Fatalf("el archivo dentro de la carpeta eliminada no debe existir")
	}
}

func TestEjecutarREMOVENoEliminaRamaSinPermisoEnHijo(t *testing.T) { // esta prueba valida que remove no haga eliminaciones parciales si un hijo no tiene permiso
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKFILE(map[string]string{"path": "/docs/sub/a.txt", "size": "5"}, map[string]bool{"r": true}); err != nil {
		t.Fatalf("mkfile devolvio error: %v", err)
	}

	if err := EjecutarMKGRP(map[string]string{"name": "usuarios"}); err != nil {
		t.Fatalf("mkgrp devolvio error: %v", err)
	}
	if err := EjecutarMKUSR(map[string]string{"user": "user1", "pass": "clave", "grp": "usuarios"}); err != nil {
		t.Fatalf("mkusr devolvio error: %v", err)
	}

	file, err := os.OpenFile(sesionActual.Montada.Path, os.O_RDWR, 0666)
	if err != nil {
		t.Fatalf("no se pudo abrir disco para permisos: %v", err)
	}
	sb := leerSuperBloquePrueba(t, file)
	docsIndex, err := buscarInodoPorRuta(file, sb, "/docs")
	if err != nil {
		t.Fatalf("no se encontro /docs: %v", err)
	}
	subIndex, err := buscarInodoPorRuta(file, sb, "/docs/sub")
	if err != nil {
		t.Fatalf("no se encontro /docs/sub: %v", err)
	}
	fileIndex, err := buscarInodoPorRuta(file, sb, "/docs/sub/a.txt")
	if err != nil {
		t.Fatalf("no se encontro /docs/sub/a.txt: %v", err)
	}

	docsInode, err := leerInodoPorIndice(file, sb, docsIndex)
	if err != nil {
		t.Fatalf("no se pudo leer /docs: %v", err)
	}
	docsInode.Perm = utils.StringToBytes3("777")
	if err := escribirInodoPorIndice(file, sb, docsIndex, docsInode); err != nil {
		t.Fatalf("no se pudo actualizar /docs: %v", err)
	}

	subInode, err := leerInodoPorIndice(file, sb, subIndex)
	if err != nil {
		t.Fatalf("no se pudo leer /docs/sub: %v", err)
	}
	subInode.Perm = utils.StringToBytes3("777")
	if err := escribirInodoPorIndice(file, sb, subIndex, subInode); err != nil {
		t.Fatalf("no se pudo actualizar /docs/sub: %v", err)
	}

	childInode, err := leerInodoPorIndice(file, sb, fileIndex)
	if err != nil {
		t.Fatalf("no se pudo leer hijo: %v", err)
	}
	childInode.Perm = utils.StringToBytes3("444")
	if err := escribirInodoPorIndice(file, sb, fileIndex, childInode); err != nil {
		t.Fatalf("no se pudo actualizar hijo: %v", err)
	}
	file.Close()

	if err := EjecutarLogout(); err != nil {
		t.Fatalf("logout devolvio error: %v", err)
	}
	if err := EjecutarLogin(map[string]string{"user": "user1", "pass": "clave", "id": "631A"}); err != nil {
		t.Fatalf("login user1 devolvio error: %v", err)
	}

	if err := EjecutarREMOVE(map[string]string{"path": "/docs/sub"}); err == nil {
		t.Fatalf("remove debe fallar porque el hijo no tiene permiso de escritura")
	}

	if contenido := leerArchivoPrueba(t, "/docs/sub/a.txt"); contenido != "01234" {
		t.Fatalf("remove no debe borrar parcialmente el hijo, contenido=%q", contenido)
	}
}

func TestEjecutarCOPYCopiaArchivo(t *testing.T) { // esta prueba valida copy sobre un archivo
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKFILE(map[string]string{"path": "/docs/a.txt", "size": "5"}, map[string]bool{"r": true}); err != nil {
		t.Fatalf("mkfile devolvio error: %v", err)
	}
	if err := EjecutarMKDIR(map[string]string{"path": "/backup"}, map[string]bool{}); err != nil {
		t.Fatalf("mkdir devolvio error: %v", err)
	}

	if err := EjecutarCOPY(map[string]string{"path": "/docs/a.txt", "destino": "/backup"}); err != nil {
		t.Fatalf("copy devolvio error: %v", err)
	}

	if contenido := leerArchivoPrueba(t, "/backup/a.txt"); contenido != "01234" {
		t.Fatalf("contenido copiado inesperado: %q", contenido)
	}
}

func TestEjecutarMOVEMueveArchivo(t *testing.T) { // esta prueba valida move cambiando referencias de carpeta
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKFILE(map[string]string{"path": "/docs/a.txt", "size": "5"}, map[string]bool{"r": true}); err != nil {
		t.Fatalf("mkfile devolvio error: %v", err)
	}
	if err := EjecutarMKDIR(map[string]string{"path": "/backup"}, map[string]bool{}); err != nil {
		t.Fatalf("mkdir devolvio error: %v", err)
	}

	if err := EjecutarMOVE(map[string]string{"path": "/docs/a.txt", "destino": "/backup"}); err != nil {
		t.Fatalf("move devolvio error: %v", err)
	}

	if contenido := leerArchivoPrueba(t, "/backup/a.txt"); contenido != "01234" {
		t.Fatalf("contenido movido inesperado: %q", contenido)
	}
	if _, err := leerArchivoPruebaConError(t, "/docs/a.txt"); err == nil {
		t.Fatalf("la ruta anterior no debe existir despues de move")
	}
}

func TestEjecutarCOPYOmitHijoSinPermisoLectura(t *testing.T) { // esta prueba valida que copy continue si un hijo no puede leerse
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKFILE(map[string]string{"path": "/docs/sub/a.txt", "size": "5"}, map[string]bool{"r": true}); err != nil {
		t.Fatalf("mkfile devolvio error: %v", err)
	}
	if err := EjecutarMKDIR(map[string]string{"path": "/backup"}, map[string]bool{}); err != nil {
		t.Fatalf("mkdir backup devolvio error: %v", err)
	}
	if err := EjecutarMKGRP(map[string]string{"name": "usuarios"}); err != nil {
		t.Fatalf("mkgrp devolvio error: %v", err)
	}
	if err := EjecutarMKUSR(map[string]string{"user": "user1", "pass": "clave", "grp": "usuarios"}); err != nil {
		t.Fatalf("mkusr devolvio error: %v", err)
	}

	file, err := os.OpenFile(sesionActual.Montada.Path, os.O_RDWR, 0666)
	if err != nil {
		t.Fatalf("no se pudo abrir disco para permisos: %v", err)
	}
	sb := leerSuperBloquePrueba(t, file)
	for _, path := range []string{"/docs", "/docs/sub", "/backup"} {
		inodeIndex, err := buscarInodoPorRuta(file, sb, path)
		if err != nil {
			t.Fatalf("no se encontro %s: %v", path, err)
		}
		inode, err := leerInodoPorIndice(file, sb, inodeIndex)
		if err != nil {
			t.Fatalf("no se pudo leer %s: %v", path, err)
		}
		inode.Perm = utils.StringToBytes3("777")
		if err := escribirInodoPorIndice(file, sb, inodeIndex, inode); err != nil {
			t.Fatalf("no se pudo actualizar %s: %v", path, err)
		}
	}
	childIndex, err := buscarInodoPorRuta(file, sb, "/docs/sub/a.txt")
	if err != nil {
		t.Fatalf("no se encontro hijo: %v", err)
	}
	childInode, err := leerInodoPorIndice(file, sb, childIndex)
	if err != nil {
		t.Fatalf("no se pudo leer hijo: %v", err)
	}
	childInode.Perm = utils.StringToBytes3("222")
	if err := escribirInodoPorIndice(file, sb, childIndex, childInode); err != nil {
		t.Fatalf("no se pudo actualizar hijo: %v", err)
	}
	file.Close()

	if err := EjecutarLogout(); err != nil {
		t.Fatalf("logout devolvio error: %v", err)
	}
	if err := EjecutarLogin(map[string]string{"user": "user1", "pass": "clave", "id": "631A"}); err != nil {
		t.Fatalf("login user1 devolvio error: %v", err)
	}

	if err := EjecutarCOPY(map[string]string{"path": "/docs", "destino": "/backup"}); err != nil {
		t.Fatalf("copy debe continuar omitiendo el hijo sin permiso: %v", err)
	}
	if _, err := leerArchivoPruebaConError(t, "/backup/docs/sub/a.txt"); err == nil {
		t.Fatalf("el hijo sin permiso de lectura no debe copiarse")
	}
}

func TestEjecutarMOVERechazaMoverCarpetaDentroDeSiMisma(t *testing.T) { // esta prueba valida proteccion contra ciclos al mover carpetas
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKDIR(map[string]string{"path": "/docs/sub"}, map[string]bool{"p": true}); err != nil {
		t.Fatalf("mkdir devolvio error: %v", err)
	}
	if err := EjecutarMOVE(map[string]string{"path": "/docs", "destino": "/docs/sub"}); err == nil {
		t.Fatalf("move debe rechazar mover una carpeta dentro de si misma")
	}
}

func leerArchivoPruebaConError(t *testing.T, path string) (string, error) {
	t.Helper()
	file, sb := abrirDiscoPrueba(t)
	defer file.Close()
	return leerArchivoPorRuta(file, sb, path)
}

func leerSuperBloquePrueba(t *testing.T, file *os.File) estructuras.SuperBlock {
	t.Helper()
	var sb estructuras.SuperBlock
	if err := utils.ReadStructAt(file, int64(sesionActual.Montada.Partition.Start), &sb); err != nil {
		t.Fatalf("no se pudo leer superbloque: %v", err)
	}
	return sb
}
