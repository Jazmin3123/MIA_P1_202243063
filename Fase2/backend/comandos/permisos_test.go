package comandos

import "testing" // uso testing para validar permisos ugo

func TestPermisosRechazanEscrituraDeOtroUsuario(t *testing.T) { // esta prueba valida que un usuario normal no escriba donde no tiene permiso
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKGRP(map[string]string{"name": "usuarios"}); err != nil {
		t.Fatalf("mkgrp devolvio error: %v", err)
	}

	if err := EjecutarMKUSR(map[string]string{"user": "user1", "pass": "clave", "grp": "usuarios"}); err != nil {
		t.Fatalf("mkusr devolvio error: %v", err)
	}

	if err := EjecutarMKDIR(map[string]string{"path": "/privado"}, map[string]bool{}); err != nil {
		t.Fatalf("mkdir devolvio error: %v", err)
	}

	if err := EjecutarLogout(); err != nil {
		t.Fatalf("logout devolvio error: %v", err)
	}

	if err := EjecutarLogin(map[string]string{"user": "user1", "pass": "clave", "id": "631A"}); err != nil {
		t.Fatalf("login devolvio error: %v", err)
	}

	if err := EjecutarMKFILE(map[string]string{"path": "/privado/a.txt", "size": "5"}, map[string]bool{}); err == nil {
		t.Fatalf("mkfile debio rechazar escritura sin permiso")
	}
}
