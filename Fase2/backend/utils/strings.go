package utils

import "strings" // uso strings para limpiar bytes vacios al convertir a texto

func StringToBytes16(text string) [16]byte { // esta funcion convierte texto a un arreglo fijo de 16 bytes
	var result [16]byte
	copy(result[:], []byte(text))
	return result
}

func StringToBytes12(text string) [12]byte { // esta funcion convierte texto a un arreglo fijo de 12 bytes
	var result [12]byte
	copy(result[:], []byte(text))
	return result
}

func StringToBytes4(text string) [4]byte { // esta funcion convierte texto a un arreglo fijo de 4 bytes
	var result [4]byte
	copy(result[:], []byte(text))
	return result
}

func StringToBytes3(text string) [3]byte { // esta funcion convierte texto a un arreglo fijo de 3 bytes
	var result [3]byte
	copy(result[:], []byte(text))
	return result
}

func BytesToString(data []byte) string { // esta funcion convierte bytes fijos a texto limpio
	return strings.TrimRight(string(data), "\x00")
}
