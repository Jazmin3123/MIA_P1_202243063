package comandos

import (
	"fmt"     // uso fmt para errores y mensajes de sesion
	"os"      // uso os para abrir el disco donde esta users.txt
	"strconv" // uso strconv para convertir ids de usuarios y grupos
	"strings" // uso strings para partir lineas del archivo users.txt

	"MIA_P1_202243063/estructuras" // uso estructuras para leer superbloque, inodos y bloques
	"MIA_P1_202243063/utils"       // uso utils para leer estructuras binarias y limpiar textos
)

type UsuarioSistema struct { // esta estructura representa un usuario leido desde users.txt
	UID      int32  // aqui guardo el id del usuario
	GID      int32  // aqui guardo el id del grupo al que pertenece
	Grupo    string // aqui guardo el nombre del grupo
	Usuario  string // aqui guardo el nombre del usuario
	Password string // aqui guardo la contrasena
	Activo   bool   // aqui guardo si el registro no esta eliminado
}

type GrupoSistema struct { // esta estructura representa un grupo leido desde users.txt
	GID    int32  // aqui guardo el id del grupo
	Nombre string // aqui guardo el nombre del grupo
	Activo bool   // aqui guardo si el grupo no esta eliminado
}

type SesionSistema struct { // esta estructura guarda la sesion activa en memoria
	Activa  bool             // aqui indico si hay una sesion iniciada
	Usuario UsuarioSistema   // aqui guardo el usuario autenticado
	Montada MountedPartition // aqui guardo la particion donde inicio sesion
}

var sesionActual SesionSistema // aqui guardo la sesion actual mientras el programa esta abierto

func EjecutarLogin(params map[string]string) error { // esta funcion inicia sesion con usuario, password e id
	if sesionActual.Activa {
		return fmt.Errorf("ya existe una sesion activa, primero ejecuta logout")
	}

	mounted, exists := FindMountedPartitionByID(params["id"]) // busco la particion montada donde se iniciara sesion
	if !exists {
		return fmt.Errorf("no existe una particion montada con id %s", params["id"])
	}

	usuarios, _, err := leerUsuariosYGrupos(*mounted) // leo users.txt desde la particion ext2
	if err != nil {
		return err
	}

	for _, usuario := range usuarios {
		if usuario.Activo && usuario.Usuario == params["user"] && usuario.Password == params["pass"] {
			sesionActual = SesionSistema{Activa: true, Usuario: usuario, Montada: *mounted}
			fmt.Printf("sesion iniciada correctamente: %s en %s\n", usuario.Usuario, mounted.ID)
			return nil
		}
	}

	return fmt.Errorf("autenticacion fallida")
}

func EjecutarLogout() error { // esta funcion cierra la sesion activa
	if !sesionActual.Activa {
		return fmt.Errorf("no existe una sesion activa")
	}

	fmt.Printf("sesion cerrada: %s\n", sesionActual.Usuario.Usuario)
	sesionActual = SesionSistema{}
	return nil
}

func leerUsuariosYGrupos(mounted MountedPartition) ([]UsuarioSistema, []GrupoSistema, error) { // esta funcion lee y parsea users.txt
	contenido, err := leerUsersTxt(mounted) // leo el contenido real del archivo users.txt
	if err != nil {
		return nil, nil, err
	}

	grupos := map[string]GrupoSistema{}
	var usuarios []UsuarioSistema

	for _, linea := range strings.Split(contenido, "\n") {
		linea = strings.TrimSpace(linea)
		if linea == "" {
			continue
		}

		partes := separarRegistroUsers(linea) // separo campos tipo csv simple
		if len(partes) < 3 {
			continue
		}

		id, err := strconv.Atoi(partes[0])
		if err != nil {
			continue
		}

		tipo := strings.ToUpper(partes[1])
		if tipo == "G" && len(partes) >= 3 {
			grupo := GrupoSistema{GID: int32(id), Nombre: partes[2], Activo: id != 0}
			grupos[grupo.Nombre] = grupo
		}

		if tipo == "U" && len(partes) >= 5 {
			grupo := grupos[partes[2]]
			usuarios = append(usuarios, UsuarioSistema{
				UID:      int32(id), // id del usuario
				GID:      grupo.GID, // id del grupo encontrado
				Grupo:    partes[2], // nombre del grupo
				Usuario:  partes[3], // nombre del usuario
				Password: partes[4], // contrasena del usuario
				Activo:   id != 0,   // id cero significa eliminado
			})
		}
	}

	var listaGrupos []GrupoSistema
	for _, grupo := range grupos {
		listaGrupos = append(listaGrupos, grupo)
	}

	return usuarios, listaGrupos, nil
}

func leerUsersTxt(mounted MountedPartition) (string, error) { // esta funcion lee el archivo /users.txt desde ext2
	file, err := os.Open(mounted.Path)
	if err != nil {
		return "", fmt.Errorf("no se pudo abrir el disco: %w", err)
	}
	defer file.Close()

	var sb estructuras.SuperBlock
	if err := utils.ReadStructAt(file, int64(mounted.Partition.Start), &sb); err != nil {
		return "", err
	}

	if sb.Magic != estructuras.Ext2Magic {
		return "", fmt.Errorf("la particion no esta formateada como ext2")
	}

	var usersInode estructuras.Inode
	if err := utils.ReadStructAt(file, int64(sb.InodeStart+sb.InodeSize), &usersInode); err != nil {
		return "", err
	}

	contenido, err := leerContenidoArchivo(file, sb, usersInode) // leo bloques directos del archivo
	if err != nil {
		return "", err
	}

	return contenido, nil
}

func leerContenidoArchivo(file *os.File, sb estructuras.SuperBlock, inode estructuras.Inode) (string, error) { // esta funcion lee contenido usando apuntadores directos
	if inode.Size == 0 {
		return "", nil
	}

	resultado := make([]byte, 0, inode.Size)
	pendiente := inode.Size

	for index := 0; index < 12 && pendiente > 0; index++ {
		blockIndex := inode.Block[index]
		if blockIndex == -1 {
			continue
		}

		var block estructuras.FileBlock
		blockPosition := int64(sb.BlockStart + blockIndex*sb.BlockSize)
		if err := utils.ReadStructAt(file, blockPosition, &block); err != nil {
			return "", err
		}

		copiar := sb.BlockSize
		if pendiente < copiar {
			copiar = pendiente
		}

		resultado = append(resultado, block.Content[:copiar]...)
		pendiente -= copiar
	}

	return utils.BytesToString(resultado), nil
}

func separarRegistroUsers(linea string) []string { // esta funcion limpia los campos separados por coma en users.txt
	partesCrudas := strings.Split(linea, ",")
	partes := make([]string, 0, len(partesCrudas))

	for _, parte := range partesCrudas {
		partes = append(partes, strings.TrimSpace(parte))
	}

	return partes
}

func obtenerSesionActual() (SesionSistema, bool) { // esta funcion devuelve la sesion para futuros comandos
	return sesionActual, sesionActual.Activa
}

func limpiarSesionParaPrueba() { // esta funcion limpia la sesion en pruebas automaticas
	sesionActual = SesionSistema{}
}
