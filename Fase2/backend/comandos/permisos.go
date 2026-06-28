package comandos

import (
	"fmt"     // uso fmt para explicar errores de permisos
	"strings" // uso strings para limpiar los permisos guardados en el inodo

	"MIA_P1_202243063/estructuras" // uso estructuras para recibir inodos ext2
	"MIA_P1_202243063/utils"       // uso utils para convertir permisos binarios a texto
)

const (
	permisoLectura   = 4 // este bit representa permiso de lectura
	permisoEscritura = 2 // este bit representa permiso de escritura
	permisoEjecucion = 1 // este bit representa permiso de ejecucion
)

func usuarioPuedeUsarInodo(inode estructuras.Inode, usuario UsuarioSistema, permiso int) bool { // esta funcion revisa permisos ugo contra el usuario activo
	if usuario.UID == 1 || strings.EqualFold(usuario.Usuario, "root") {
		return true
	}

	permisos := utils.BytesToString(inode.Perm[:])
	if len(permisos) != 3 {
		return false
	}

	posicion := 2
	if usuario.UID == inode.UID {
		posicion = 0
	} else if usuario.GID == inode.GID {
		posicion = 1
	}

	digito := int(permisos[posicion] - '0')
	if digito < 0 || digito > 7 {
		return false
	}

	return digito&permiso == permiso
}

func validarPermisoInodo(inode estructuras.Inode, usuario UsuarioSistema, permiso int, accion string) error { // esta funcion devuelve error cuando el usuario no tiene permiso suficiente
	if usuarioPuedeUsarInodo(inode, usuario, permiso) {
		return nil
	}

	return fmt.Errorf("permiso denegado para %s", accion)
}
