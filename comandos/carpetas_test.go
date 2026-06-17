package comandos

import (
	"os"      // uso os para abrir el disco de prueba
	"testing" // uso testing para validar mkdir

	"MIA_P1_202243063/estructuras" // uso estructuras para leer superbloque
	"MIA_P1_202243063/utils"       // uso utils para leer structs binarios
)

func TestEjecutarMKDIRCreaRutaConPadres(t *testing.T) { // esta prueba valida mkdir -p dentro de ext2
	prepararSesionRootParaUsuarios(t)

	if err := EjecutarMKDIR(map[string]string{"path": "/home/user/docs"}, map[string]bool{"p": true}); err != nil {
		t.Fatalf("mkdir devolvio error: %v", err)
	}

	file, err := os.Open(sesionActual.Montada.Path)
	if err != nil {
		t.Fatalf("no se pudo abrir el disco: %v", err)
	}
	defer file.Close()

	var sb estructuras.SuperBlock
	if err := utils.ReadStructAt(file, int64(sesionActual.Montada.Partition.Start), &sb); err != nil {
		t.Fatalf("no se pudo leer superbloque: %v", err)
	}

	home, exists, err := buscarEntradaEnCarpeta(file, sb, 0, "home")
	if err != nil || !exists {
		t.Fatalf("no se encontro /home: exists=%v err=%v", exists, err)
	}

	user, exists, err := buscarEntradaEnCarpeta(file, sb, home, "user")
	if err != nil || !exists {
		t.Fatalf("no se encontro /home/user: exists=%v err=%v", exists, err)
	}

	_, exists, err = buscarEntradaEnCarpeta(file, sb, user, "docs")
	if err != nil || !exists {
		t.Fatalf("no se encontro /home/user/docs: exists=%v err=%v", exists, err)
	}
}

func TestEjecutarMKDIRRequierePadresOSesion(t *testing.T) { // esta prueba valida errores basicos de mkdir
	limpiarSesionParaPrueba()
	if err := EjecutarMKDIR(map[string]string{"path": "/home"}, map[string]bool{}); err == nil {
		t.Fatalf("mkdir debe requerir sesion activa")
	}

	prepararSesionRootParaUsuarios(t)
	if err := EjecutarMKDIR(map[string]string{"path": "/a/b"}, map[string]bool{}); err == nil {
		t.Fatalf("mkdir sin -p debe fallar si no existe el padre")
	}
}
