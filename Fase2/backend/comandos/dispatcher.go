package comandos

import (
	"fmt" // uso fmt para mostrar mensajes al usuario

	"MIA_P1_202243063/utils" // uso utils para recibir la estructura Command
)

func ExecuteCommand(cmd utils.Command) { // esta funcion decide que hacer con el comando ya validado para la CLI
	if err := ExecuteCommandResult(cmd); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func ExecuteCommandResult(cmd utils.Command) error { // esta funcion ejecuta comandos y devuelve errores para reutilizar en API
	if cmd.Name == "execute" {
		return runScript(cmd.Params["path"]) // execute carga y procesa un archivo .smia
	}

	if cmd.Name == "mkdisk" {
		return ExecuteMKDisk(cmd.Params) // ejecuto la creacion real del disco
	}

	if cmd.Name == "rmdisk" {
		return ExecuteRMDisk(cmd.Params) // ejecuto la eliminacion real del disco
	}

	if cmd.Name == "fdisk" {
		return ExecuteFDisk(cmd.Params) // ejecuto la creacion real de particiones
	}

	if cmd.Name == "mount" {
		return ExecuteMount(cmd.Params) // ejecuto el montaje en memoria de la particion
	}

	if cmd.Name == "mkfs" {
		return ExecuteMKFS(cmd.Params) // ejecuto el formateo ext2 de la particion montada
	}

	if cmd.Name == "login" {
		return EjecutarLogin(cmd.Params) // ejecuto el inicio de sesion sobre una particion montada
	}

	if cmd.Name == "logout" {
		return EjecutarLogout() // ejecuto el cierre de la sesion actual
	}

	if cmd.Name == "mkgrp" {
		return EjecutarMKGRP(cmd.Params) // ejecuto la creacion de grupos en users.txt
	}

	if cmd.Name == "rmgrp" {
		return EjecutarRMGRP(cmd.Params) // ejecuto la eliminacion logica de grupos
	}

	if cmd.Name == "mkusr" {
		return EjecutarMKUSR(cmd.Params) // ejecuto la creacion de usuarios en users.txt
	}

	if cmd.Name == "rmusr" {
		return EjecutarRMUSR(cmd.Params) // ejecuto la eliminacion logica de usuarios
	}

	if cmd.Name == "chgrp" {
		return EjecutarCHGRP(cmd.Params) // ejecuto el cambio de grupo de un usuario
	}

	if cmd.Name == "pause" {
		EjecutarPause() // detengo la ejecucion hasta presionar enter
		return nil
	}

	if cmd.Name == "mkdir" {
		return EjecutarMKDIR(cmd.Params, cmd.Flags) // ejecuto la creacion de carpetas dentro de ext2
	}

	if cmd.Name == "mkfile" {
		return EjecutarMKFILE(cmd.Params, cmd.Flags) // ejecuto la creacion de archivos dentro de ext2
	}

	if cmd.Name == "cat" {
		return EjecutarCAT(cmd.Params) // ejecuto la lectura de archivos dentro de ext2
	}

	if cmd.Name == "rename" {
		return EjecutarRENAME(cmd.Params) // ejecuto el renombrado de archivos o carpetas
	}

	if cmd.Name == "edit" {
		return EjecutarEDIT(cmd.Params) // ejecuto la edicion de archivos dentro de ext2
	}

	if cmd.Name == "remove" {
		return EjecutarREMOVE(cmd.Params) // ejecuto la eliminacion recursiva dentro de ext2
	}

	if cmd.Name == "copy" {
		return EjecutarCOPY(cmd.Params) // ejecuto la copia recursiva dentro de ext2
	}

	if cmd.Name == "move" {
		return EjecutarMOVE(cmd.Params) // ejecuto el movimiento dentro de ext2
	}

	if cmd.Name == "rep" {
		return EjecutarREP(cmd.Params) // ejecuto la generacion de reportes
	}

	fmt.Printf("Comando reconocido: %s\n", cmd.Name) // por ahora solo confirmo que el comando fue valido

	if len(cmd.Params) > 0 {
		fmt.Println("Parametros:")
		for key, value := range cmd.Params { // muestro parametros con valor para verificar el parser
			fmt.Printf("  -%s=%s\n", key, value)
		}
	}

	if len(cmd.Flags) > 0 {
		fmt.Println("Banderas:")
		for key := range cmd.Flags { // muestro banderas sin valor para verificar el parser
			fmt.Printf("  -%s\n", key)
		}
	}

	return nil
}

func EjecutarPause() { // esta funcion implementa el comando pause del pdf actualizado
	fmt.Print("pause: presiona enter para continuar...")
	_, _ = consoleReader.ReadString('\n')
}
