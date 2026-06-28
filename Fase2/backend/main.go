package main

import (
	"fmt"
	"os"

	"MIA_P1_202243063/api"
	"MIA_P1_202243063/comandos" // importo el paquete donde deje la consola del proyecto
)

func main() { // aqui empieza la ejecucion del programa
	if len(os.Args) > 1 && (os.Args[1] == "api" || os.Args[1] == "--api") {
		fmt.Println("API escuchando en http://localhost:8080")
		if err := api.StartAPI(); err != nil {
			fmt.Println("Error iniciando API:", err)
		}
		return
	}

	comandos.RunCLI() // llamo a la cli para que el programa empiece a recibir comandos
}
