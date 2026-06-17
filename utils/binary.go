package utils

import (
	"encoding/binary" // uso binary para escribir structs en formato binario
	"fmt"             // uso fmt para devolver errores con contexto
	"os"              // uso os para trabajar con archivos abiertos
)

var ByteOrder = binary.LittleEndian // aqui defino el orden de bytes que usare en todo el proyecto

func WriteStructAt(file *os.File, position int64, data any) error { // esta funcion escribe una estructura en una posicion exacta del archivo
	if _, err := file.Seek(position, 0); err != nil {
		return fmt.Errorf("no se pudo mover a la posicion %d: %w", position, err)
	}

	if err := binary.Write(file, ByteOrder, data); err != nil {
		return fmt.Errorf("no se pudo escribir la estructura: %w", err)
	}

	return nil
}

func ReadStructAt(file *os.File, position int64, data any) error { // esta funcion lee una estructura desde una posicion exacta del archivo
	if _, err := file.Seek(position, 0); err != nil {
		return fmt.Errorf("no se pudo mover a la posicion %d: %w", position, err)
	}

	if err := binary.Read(file, ByteOrder, data); err != nil {
		return fmt.Errorf("no se pudo leer la estructura: %w", err)
	}

	return nil
}

func WriteBytesAt(file *os.File, position int64, data []byte) error { // esta funcion escribe bytes crudos en una posicion exacta
	if _, err := file.Seek(position, 0); err != nil {
		return fmt.Errorf("no se pudo mover a la posicion %d: %w", position, err)
	}

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("no se pudieron escribir bytes: %w", err)
	}

	return nil
}

func StructSize(data any) int32 { // esta funcion devuelve cuantos bytes ocupa una estructura binaria
	return int32(binary.Size(data))
}
