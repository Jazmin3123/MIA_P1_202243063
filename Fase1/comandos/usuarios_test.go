package comandos

import (
	"strings" // uso strings para validar el contenido de users.txt
	"testing" // uso testing para validar usuarios y grupos
)

func TestAdministracionUsuariosYGruposRoot(t *testing.T) { // esta prueba valida mkgrp, mkusr, chgrp, rmusr y rmgrp
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKGRP(map[string]string{"name": "usuarios"}); err != nil {
		t.Fatalf("mkgrp devolvio error: %v", err)
	}

	if err := EjecutarMKUSR(map[string]string{"user": "user1", "pass": "clave", "grp": "usuarios"}); err != nil {
		t.Fatalf("mkusr devolvio error: %v", err)
	}

	contenido, err := leerUsersTxt(sesionActual.Montada)
	if err != nil {
		t.Fatalf("no se pudo leer users.txt: %v", err)
	}

	if !strings.Contains(contenido, "2,G,usuarios") {
		t.Fatalf("users.txt no contiene el grupo creado: %q", contenido)
	}

	if !strings.Contains(contenido, "2,U,usuarios,user1,clave") {
		t.Fatalf("users.txt no contiene el usuario creado: %q", contenido)
	}

	if err := EjecutarMKGRP(map[string]string{"name": "devs"}); err != nil {
		t.Fatalf("mkgrp devs devolvio error: %v", err)
	}

	if err := EjecutarCHGRP(map[string]string{"user": "user1", "grp": "devs"}); err != nil {
		t.Fatalf("chgrp devolvio error: %v", err)
	}

	contenido, _ = leerUsersTxt(sesionActual.Montada)
	if !strings.Contains(contenido, "2,U,devs,user1,clave") {
		t.Fatalf("users.txt no cambio el grupo del usuario: %q", contenido)
	}

	if err := EjecutarRMUSR(map[string]string{"user": "user1"}); err != nil {
		t.Fatalf("rmusr devolvio error: %v", err)
	}

	if err := EjecutarRMGRP(map[string]string{"name": "usuarios"}); err != nil {
		t.Fatalf("rmgrp devolvio error: %v", err)
	}

	contenido, _ = leerUsersTxt(sesionActual.Montada)
	if !strings.Contains(contenido, "0,U,devs,user1,clave") {
		t.Fatalf("users.txt no marco usuario eliminado: %q", contenido)
	}

	if !strings.Contains(contenido, "0,G,usuarios") {
		t.Fatalf("users.txt no marco grupo eliminado: %q", contenido)
	}
}

func TestAdministracionUsuariosRequiereRoot(t *testing.T) { // esta prueba valida que solo root pueda administrar usuarios
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKGRP(map[string]string{"name": "usuarios"}); err != nil {
		t.Fatalf("mkgrp devolvio error: %v", err)
	}

	if err := EjecutarMKUSR(map[string]string{"user": "user1", "pass": "clave", "grp": "usuarios"}); err != nil {
		t.Fatalf("mkusr devolvio error: %v", err)
	}

	if err := EjecutarLogout(); err != nil {
		t.Fatalf("logout devolvio error: %v", err)
	}

	if err := EjecutarLogin(map[string]string{"user": "user1", "pass": "clave", "id": "631A"}); err != nil {
		t.Fatalf("login user1 devolvio error: %v", err)
	}

	if err := EjecutarMKGRP(map[string]string{"name": "otro"}); err == nil {
		t.Fatalf("mkgrp debe rechazar usuarios que no sean root")
	}
}

func prepararSesionRootParaUsuarios(t *testing.T) { // esta funcion prepara una sesion root sobre ext2
	t.Helper()
	prepararParticionFormateadaParaSesion(t)

	if err := EjecutarLogin(map[string]string{"user": "root", "pass": "123", "id": "631A"}); err != nil {
		t.Fatalf("login root devolvio error: %v", err)
	}
}
