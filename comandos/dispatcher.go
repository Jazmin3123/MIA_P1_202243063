package comandos

import (
	"fmt" // uso fmt para mostrar mensajes al usuario

	"MIA_P1_202243063/utils" // uso utils para recibir la estructura Command
)

func ExecuteCommand(cmd utils.Command) { // esta funcion decide que hacer con el comando ya validado
	if cmd.Name == "execute" {
		if err := runScript(cmd.Params["path"]); err != nil { // execute carga y procesa un archivo .smia
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if cmd.Name == "mkdisk" {
		if err := ExecuteMKDisk(cmd.Params); err != nil { // ejecuto la creacion real del disco
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if cmd.Name == "rmdisk" {
		if err := ExecuteRMDisk(cmd.Params); err != nil { // ejecuto la eliminacion real del disco
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if cmd.Name == "fdisk" {
		if err := ExecuteFDisk(cmd.Params); err != nil { // ejecuto la creacion real de particiones
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if cmd.Name == "mount" {
		if err := ExecuteMount(cmd.Params); err != nil { // ejecuto el montaje en memoria de la particion
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if cmd.Name == "mkfs" {
		if err := ExecuteMKFS(cmd.Params); err != nil { // ejecuto el formateo ext2 de la particion montada
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if cmd.Name == "login" {
		if err := EjecutarLogin(cmd.Params); err != nil { // ejecuto el inicio de sesion sobre una particion montada
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if cmd.Name == "logout" {
		if err := EjecutarLogout(); err != nil { // ejecuto el cierre de la sesion actual
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if cmd.Name == "mkgrp" {
		if err := EjecutarMKGRP(cmd.Params); err != nil { // ejecuto la creacion de grupos en users.txt
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if cmd.Name == "rmgrp" {
		if err := EjecutarRMGRP(cmd.Params); err != nil { // ejecuto la eliminacion logica de grupos
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if cmd.Name == "mkusr" {
		if err := EjecutarMKUSR(cmd.Params); err != nil { // ejecuto la creacion de usuarios en users.txt
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if cmd.Name == "rmusr" {
		if err := EjecutarRMUSR(cmd.Params); err != nil { // ejecuto la eliminacion logica de usuarios
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if cmd.Name == "chgrp" {
		if err := EjecutarCHGRP(cmd.Params); err != nil { // ejecuto el cambio de grupo de un usuario
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if cmd.Name == "pause" {
		EjecutarPause() // detengo la ejecucion hasta presionar enter
		return
	}

	if cmd.Name == "mkdir" {
		if err := EjecutarMKDIR(cmd.Params, cmd.Flags); err != nil { // ejecuto la creacion de carpetas dentro de ext2
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if cmd.Name == "mkfile" {
		if err := EjecutarMKFILE(cmd.Params, cmd.Flags); err != nil { // ejecuto la creacion de archivos dentro de ext2
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if cmd.Name == "cat" {
		if err := EjecutarCAT(cmd.Params); err != nil { // ejecuto la lectura de archivos dentro de ext2
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if cmd.Name == "rep" {
		if err := EjecutarREP(cmd.Params); err != nil { // ejecuto la generacion de reportes
			fmt.Printf("Error: %v\n", err)
		}
		return
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
}

func EjecutarPause() { // esta funcion implementa el comando pause del pdf actualizado
	fmt.Print("pause: presiona enter para continuar...")
	_, _ = consoleReader.ReadString('\n')
}
