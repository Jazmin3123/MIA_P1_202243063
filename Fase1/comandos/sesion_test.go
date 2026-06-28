package comandos

import "testing" // uso testing para validar login y logout

func TestEjecutarLoginYLogoutRoot(t *testing.T) { // esta prueba valida login y logout con el usuario root inicial
	prepararParticionFormateadaParaSesion(t)

	if err := EjecutarLogin(map[string]string{"user": "root", "pass": "123", "id": "631A"}); err != nil {
		t.Fatalf("login root devolvio error: %v", err)
	}

	sesion, activa := obtenerSesionActual()
	if !activa {
		t.Fatalf("debe existir una sesion activa")
	}

	if sesion.Usuario.Usuario != "root" {
		t.Fatalf("el usuario activo debe ser root, fue %s", sesion.Usuario.Usuario)
	}

	if err := EjecutarLogout(); err != nil {
		t.Fatalf("logout devolvio error: %v", err)
	}

	if _, activa := obtenerSesionActual(); activa {
		t.Fatalf("la sesion debe quedar cerrada")
	}
}

func TestEjecutarLoginRechazaCredencialesInvalidas(t *testing.T) { // esta prueba valida que login rechace password incorrecta
	prepararParticionFormateadaParaSesion(t)

	if err := EjecutarLogin(map[string]string{"user": "root", "pass": "incorrecta", "id": "631A"}); err == nil {
		t.Fatalf("login debe rechazar credenciales invalidas")
	}
}

func TestEjecutarLoginRechazaSesionDuplicada(t *testing.T) { // esta prueba valida que no se pueda abrir otra sesion encima
	prepararParticionFormateadaParaSesion(t)

	if err := EjecutarLogin(map[string]string{"user": "root", "pass": "123", "id": "631A"}); err != nil {
		t.Fatalf("login root devolvio error: %v", err)
	}

	if err := EjecutarLogin(map[string]string{"user": "root", "pass": "123", "id": "631A"}); err == nil {
		t.Fatalf("login debe rechazar una segunda sesion activa")
	}
}

func prepararParticionFormateadaParaSesion(t *testing.T) string { // esta funcion deja lista una particion ext2 montada
	t.Helper()
	resetMountsForTest()
	limpiarSesionParaPrueba()
	path := createTestDisk(t)

	if err := ExecuteFDisk(map[string]string{"size": "500", "unit": "K", "path": path, "name": "primaria1", "type": "P"}); err != nil {
		t.Fatalf("no se pudo crear primaria: %v", err)
	}

	if err := ExecuteMount(map[string]string{"path": path, "name": "primaria1"}); err != nil {
		t.Fatalf("no se pudo montar primaria: %v", err)
	}

	if err := ExecuteMKFS(map[string]string{"id": "631A"}); err != nil {
		t.Fatalf("no se pudo formatear primaria: %v", err)
	}

	return path
}
