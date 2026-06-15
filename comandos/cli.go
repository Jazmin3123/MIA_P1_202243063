package comandos

import (
	"bufio"   // uso bufio para leer comandos desde consola o archivos
	"fmt"     // uso fmt para mostrar mensajes al usuario
	"os"      // uso os para leer stdin y abrir scripts
	"strings" // uso strings para limpiar lineas y validar extensiones

	"MIA_P1_202243063/utils" // uso utils para parsear y validar comandos
)

var consoleReader = bufio.NewReader(os.Stdin) // aqui dejo un solo lector para comandos y confirmaciones

func RunCLI() { // esta funcion arranca la consola principal del proyecto
	fmt.Println("MIA Proyecto 1 - CLI")                    // mensaje inicial del programa
	fmt.Println("Escribe un comando o 'exit' para salir.") // aviso para saber como cerrar la consola

	runInteractive() // aqui inicio el ciclo que queda esperando comandos
}

func runInteractive() { // esta funcion mantiene la consola activa
	for {
		fmt.Print("MIA> ") // prompt para saber que el programa esta esperando un comando

		line, err := consoleReader.ReadString('\n') // leo una linea completa desde la consola
		if err != nil {
			fmt.Println() // salto de linea cuando se termina la entrada
			return
		}

		shouldExit := processLine(line, false) // proceso la linea como comando normal, no como script
		if shouldExit {
			return // si processline devuelve true, cierro el programa
		}
	}
}

func processLine(line string, fromScript bool) bool { // esta funcion procesa una linea escrita o leida desde script
	line = strings.TrimSpace(line) // quito espacios al inicio y al final para evitar errores

	if line == "" {
		return false // si la linea esta vacia no hago nada
	}

	if strings.HasPrefix(line, "#") {
		fmt.Println(line) // si es comentario de script lo muestro tal como viene
		return false
	}

	if strings.EqualFold(line, "exit") {
		return true // exit sirve para cerrar la consola
	}

	if fromScript {
		fmt.Println(line) // cuando viene de script, muestro el comando que se esta ejecutando
	}

	cmd, err := utils.ParseCommand(line) // convierto el texto en una estructura del comando
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}

	if err := utils.ValidateCommand(cmd); err != nil { // reviso que el comando exista y tenga parametros correctos
		fmt.Printf("Error: %v\n", err)
		return false
	}

	ExecuteCommand(cmd) // mando el comando al despachador para ejecutarlo
	return false
}

func runScript(path string) error { // esta funcion ejecuta linea por linea un archivo .smia
	if !strings.HasSuffix(strings.ToLower(path), ".smia") {
		return fmt.Errorf("el script debe tener extension .smia") // valido que sea el tipo de script pedido
	}

	file, err := os.Open(path) // abro el archivo de comandos
	if err != nil {
		return fmt.Errorf("no se pudo abrir el script: %w", err)
	}
	defer file.Close() // cierro el archivo al terminar la funcion

	scanner := bufio.NewScanner(file) // leo el script linea por linea
	lineNumber := 1                   // contador para mostrar en que linea va el script

	for scanner.Scan() {
		fmt.Printf("[%d] ", lineNumber)                 // imprimo el numero de linea del script
		shouldExit := processLine(scanner.Text(), true) // proceso la linea como si fuera escrita en consola
		if shouldExit {
			break
		}
		lineNumber++ // avanzo al siguiente numero de linea
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error leyendo el script: %w", err)
	}

	return nil
}
