package utils

import (
	"errors"  // uso errors para crear mensajes de error simples
	"fmt"     // uso fmt para formar mensajes de error con datos
	"strings" // uso strings para comparar, limpiar y separar textos
)

type Command struct { // esta estructura representa un comando despues de leerlo y separarlo
	Name   string            // aqui guardo el nombre del comando ya convertido a minusculas
	Params map[string]string // aqui guardo parametros que si traen valor, por ejemplo -path=/home/disco.mia
	Flags  map[string]bool   // aqui guardo banderas sin valor, por ejemplo -r o -p
}

var commandParams = map[string]map[string]bool{ // aqui defino que parametros acepta cada comando del enunciado
	"mkdisk":  {"size": true, "fit": true, "unit": true, "path": true},                             // parametros validos para crear discos
	"rmdisk":  {"path": true},                                                                      // parametro valido para eliminar discos
	"fdisk":   {"size": true, "unit": true, "path": true, "type": true, "fit": true, "name": true}, // parametros validos para particiones
	"mount":   {"path": true, "name": true},                                                        // parametros validos para montar particiones
	"mkfs":    {"id": true, "type": true},                                                          // parametros validos para formatear
	"cat":     {},                                                                                  // cat se valida aparte porque usa file1, file2, etc.
	"login":   {"user": true, "pass": true, "id": true},                                            // parametros validos para iniciar sesion
	"logout":  {},                                                                                  // logout no recibe parametros
	"mkgrp":   {"name": true},                                                                      // parametro valido para crear grupo
	"rmgrp":   {"name": true},                                                                      // parametro valido para eliminar grupo
	"mkusr":   {"user": true, "pass": true, "grp": true},                                           // parametros validos para crear usuario
	"rmusr":   {"user": true},                                                                      // parametro valido para eliminar usuario
	"chgrp":   {"user": true, "grp": true},                                                         // parametros validos para cambiar grupo de usuario
	"mkfile":  {"path": true, "r": true, "size": true, "cont": true},                               // parametros validos para crear archivo
	"mkdir":   {"path": true, "p": true},                                                           // parametros validos para crear carpeta
	"rep":     {"name": true, "path": true, "id": true, "path_file_ls": true},                      // parametros validos para reportes
	"execute": {"path": true},                                                                      // parametro valido para ejecutar scripts
	"pause":   {},                                                                                  // pause no recibe parametros
}

var requiredParams = map[string][]string{ // aqui indico que parametros son obligatorios para validar antes de ejecutar
	"mkdisk":  {"size", "path"},        // mkdisk necesita tamano y ruta
	"rmdisk":  {"path"},                // rmdisk necesita la ruta del disco
	"fdisk":   {"path", "name"},        // fdisk necesita disco y nombre de particion
	"mount":   {"path", "name"},        // mount necesita disco y particion
	"mkfs":    {"id"},                  // mkfs necesita el id de montaje
	"login":   {"user", "pass", "id"},  // login necesita usuario, password e id
	"mkgrp":   {"name"},                // mkgrp necesita nombre de grupo
	"rmgrp":   {"name"},                // rmgrp necesita nombre de grupo
	"mkusr":   {"user", "pass", "grp"}, // mkusr necesita usuario, password y grupo
	"rmusr":   {"user"},                // rmusr necesita nombre de usuario
	"chgrp":   {"user", "grp"},         // chgrp necesita usuario y grupo nuevo
	"mkfile":  {"path"},                // mkfile necesita ruta del archivo
	"mkdir":   {"path"},                // mkdir necesita ruta de carpeta
	"rep":     {"name", "path", "id"},  // rep necesita nombre, ruta e id
	"execute": {"path"},                // execute necesita ruta del script
}

func ParseCommand(line string) (Command, error) { // esta funcion convierte una linea de texto en un comando ordenado
	tokens, err := splitTokens(line) // separo la linea por espacios, respetando comillas
	if err != nil {
		return Command{}, err
	}

	if len(tokens) == 0 {
		return Command{}, errors.New("no se ingreso ningun comando")
	}

	cmd := Command{
		Name:   strings.ToLower(tokens[0]), // el nombre del comando no distingue mayusculas
		Params: make(map[string]string),    // inicializo el mapa de parametros con valor
		Flags:  make(map[string]bool),      // inicializo el mapa de banderas sin valor
	}

	for _, token := range tokens[1:] { // reviso cada parametro escrito despues del comando
		if !strings.HasPrefix(token, "-") {
			return Command{}, fmt.Errorf("parametro invalido %q: debe iniciar con '-'", token)
		}

		rawParam := strings.TrimPrefix(token, "-")         // quito el guion inicial
		key, value, hasValue := strings.Cut(rawParam, "=") // separo nombre y valor cuando existe =
		key = strings.ToLower(strings.TrimSpace(key))      // el parametro tambien se trabaja en minusculas

		if key == "" {
			return Command{}, fmt.Errorf("parametro invalido %q", token)
		}

		if hasValue {
			cmd.Params[key] = value // ejemplo: -size=10 queda como size -> 10
		} else {
			cmd.Flags[key] = true // ejemplo: -r queda como bandera activa
		}
	}

	return cmd, nil
}

func splitTokens(line string) ([]string, error) { // esta funcion separa la linea sin romper textos entre comillas
	var tokens []string         // aqui guardo cada parte separada de la linea
	var current strings.Builder // aqui voy formando el token actual
	inQuotes := false           // aqui controlo si estoy dentro de comillas

	for _, char := range line {
		switch char {
		case '"':
			inQuotes = !inQuotes // al encontrar comillas cambio el estado
		case ' ', '\t':
			if inQuotes {
				current.WriteRune(char) // si estoy entre comillas, el espacio forma parte del valor
				continue
			}

			if current.Len() > 0 {
				tokens = append(tokens, current.String()) // guardo el token cuando termina
				current.Reset()                           // limpio para empezar el siguiente token
			}
		default:
			current.WriteRune(char) // agrego cualquier otro caracter al token actual
		}
	}

	if inQuotes {
		return nil, errors.New("faltan comillas de cierre") // error si abri comillas y nunca las cerre
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String()) // guardo el ultimo token de la linea
	}

	return tokens, nil
}

func ValidateCommand(cmd Command) error { // esta funcion valida que el comando cumpla las reglas basicas
	allowedParams, exists := commandParams[cmd.Name] // busco los parametros permitidos del comando
	if !exists {
		return fmt.Errorf("comando no reconocido: %s", cmd.Name)
	}

	for param := range cmd.Params { // valido parametros que tienen valor
		if !isAllowedParam(cmd.Name, param, allowedParams) {
			return fmt.Errorf("parametro no reconocido para %s: -%s", cmd.Name, param)
		}
	}

	for flag := range cmd.Flags { // valido banderas sin valor
		if !isAllowedParam(cmd.Name, flag, allowedParams) {
			return fmt.Errorf("parametro no reconocido para %s: -%s", cmd.Name, flag)
		}
	}

	for _, param := range requiredParams[cmd.Name] { // reviso que no falte ningun parametro obligatorio
		if _, ok := cmd.Params[param]; !ok {
			return fmt.Errorf("falta parametro obligatorio -%s en %s", param, cmd.Name)
		}
	}

	if err := validateFlagsWithoutValue(cmd); err != nil { // reviso banderas usadas sin valor
		return err
	}

	if err := validateFlagsWithValue(cmd); err != nil { // reviso que -r y -p no vengan como -r=true
		return err
	}

	if cmd.Name == "cat" {
		return validateCatFiles(cmd) // cat necesita al menos un archivo
	}

	if cmd.Name == "fdisk" {
		return validateFDiskMode(cmd) // fdisk necesita size cuando se crea particion
	}

	if cmd.Name == "rep" {
		return validateReportName(cmd) // rep solo acepta los nombres del enunciado
	}

	return nil
}

func isAllowedParam(commandName, param string, allowedParams map[string]bool) bool { // esta funcion revisa si un parametro pertenece al comando
	if commandName == "cat" && strings.HasPrefix(param, "file") {
		return true // cat acepta file1, file2, file3, etc.
	}

	return allowedParams[param] // para los demas comandos uso la lista normal de parametros
}

func validateFlagsWithoutValue(cmd Command) error { // esta funcion valida banderas escritas sin signo igual
	validFlags := map[string]map[string]bool{ // aqui indico que banderas si pueden ir sin valor
		"mkfile": {"r": true},
		"mkdir":  {"p": true},
	}

	for flag := range cmd.Flags { // recorro las banderas que escribio el usuario
		if !validFlags[cmd.Name][flag] {
			return fmt.Errorf("el parametro -%s requiere un valor", flag)
		}
	}

	return nil
}

func validateFlagsWithValue(cmd Command) error { // esta funcion detecta banderas que no deberian traer valor
	noValueFlags := map[string]map[string]bool{ // estas banderas no deben venir con =
		"mkfile": {"r": true},
		"mkdir":  {"p": true},
	}

	for param := range cmd.Params { // reviso parametros con valor para detectar mal uso de banderas
		if noValueFlags[cmd.Name][param] {
			return fmt.Errorf("el parametro -%s no debe recibir valor", param)
		}
	}

	return nil
}

func validateCatFiles(cmd Command) error { // esta funcion valida que cat tenga al menos un archivo
	for param := range cmd.Params { // cat puede recibir varios archivos como -file1, -file2, etc.
		if strings.HasPrefix(param, "file") {
			return nil
		}
	}

	return errors.New("cat requiere al menos un parametro -fileN")
}

func validateFDiskMode(cmd Command) error { // esta funcion valida fdisk como creacion de particion en esta etapa
	if _, hasSize := cmd.Params["size"]; !hasSize {
		return errors.New("falta parametro obligatorio -size en fdisk para crear particiones") // para esta etapa fdisk se valida como creacion
	}

	return nil
}

func validateReportName(cmd Command) error { // esta funcion valida los reportes
	reportes := map[string]bool{
		"mbr":      true,
		"disk":     true,
		"inode":    true,
		"block":    true,
		"bm_inode": true,
		"bm_block": true,
		"sb":       true,
		"file":     true,
		"ls":       true,
		"tree":     true,
	}

	name := strings.ToLower(cmd.Params["name"])
	if !reportes[name] {
		return fmt.Errorf("reporte no reconocido: %s", cmd.Params["name"])
	}

	if (name == "file" || name == "ls") && cmd.Params["path_file_ls"] == "" {
		return fmt.Errorf("el reporte %s requiere -path_file_ls", name)
	}

	return nil
}
