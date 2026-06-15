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
