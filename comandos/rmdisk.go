package comandos

import (
	"fmt"     // uso fmt para mostrar preguntas y errores
	"os"      // uso os para validar y eliminar el archivo
	"strings" // uso strings para limpiar la respuesta del usuario
)

func ExecuteRMDisk(params map[string]string) error { // esta funcion ejecuta el comando rmdisk
	path := params["path"] // obtengo la ruta del disco a eliminar

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("el disco no existe: %s", path)
		}
		return fmt.Errorf("no se pudo revisar el disco: %w", err)
	}

	if !confirmDiskRemoval(path) {
		fmt.Println("eliminacion cancelada")
		return nil
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("no se pudo eliminar el disco: %w", err)
	}

	fmt.Printf("disco eliminado correctamente: %s\n", path)
	return nil
}

func confirmDiskRemoval(path string) bool { // esta funcion pide confirmacion antes de borrar el disco
	fmt.Printf("seguro que deseas eliminar el disco %s? [s/n]: ", path)

	answer, err := consoleReader.ReadString('\n') // leo la respuesta con el mismo lector de la consola
	if err != nil {
		return false
	}

	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "s" || answer == "si" || answer == "y" || answer == "yes"
}
