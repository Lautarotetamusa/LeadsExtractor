package whatsapp

import (
	"fmt"
	"slices"
	"strings"
)

// Estos son los caracteres no printables
// Se codifica un string mediante estos caracteres, tenemos 5**n bits de informacion
var RUNES = []rune{'\u200a', '\u200b', '\u200c', '\u200d', '\u200e'}

const RUNE_LEN = 3

const baseLinkUrl = "https://wa.me/%s?text=%s%s"

func encodeString(s string) string {
	var encoded strings.Builder
	for _, r := range s {
		encodedChar := encode(int(r))
		// Restamos uno porque len(encodedChar) nunca será menor a 1.
		// entonces encodeamos len=1 como len=0. Esto nos da 5 veces mas combinaciones
		length := (len(encodedChar) - 1) / RUNE_LEN

		// Añade prefijo indicando la longitud
		// El prefijo ocupa RUNE_LEN caracteres
		encoded.WriteString(encode(length))
		encoded.WriteString(encodedChar)
	}
	return encoded.String()
}

// Decodifica una cadena de caracteres no imprimibles a una cadena de caracteres imprimibles.
func decodeString(s string) (string, error) {
	var decoded strings.Builder
	i := 0

	for i < len(s) {
		// Decodifica el prefijo para obtener la longitud del carácter codificado
		if i+RUNE_LEN > len(s) {
			return "", fmt.Errorf("decoded string out of index")
		}

		lengthStr := s[i : i+RUNE_LEN]
		length := decode(lengthStr)
		if length == -1 {
			return "", fmt.Errorf("error decoding runes %q", lengthStr)
		}
		i += RUNE_LEN

		strLen := (length + 1) * RUNE_LEN
		if i+strLen > len(s) {
			return "", fmt.Errorf("decoded string out of index")
		}
		encodedChar := s[i : i+strLen]
		decodedChar := decode(encodedChar)
		if decodedChar == -1 {
			return "", fmt.Errorf("error decoding runes %q", encodedChar)
		}
		decoded.WriteRune(rune(decodedChar))
		i += strLen
	}

	return decoded.String(), nil
}

func encode(num int) string {
	base := len(RUNES)
	var output strings.Builder

	for num > 0 {
		output.WriteString(string(RUNES[num%base]))
		num /= base
	}

	// Invertir la cadena
	runes := []rune(output.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func decode(s string) int {
	base := len(RUNES)
	output := 0

	for _, char := range s {
		idx := -1
		for j, np := range RUNES {
			if np == char {
				idx = j
				break
			}
		}
		if idx == -1 {
			return -1 // Error: caracter no encontrado
		}
		output = output*base + idx
	}

	return output
}

func filterNonPrintables(msg string) string {
	var result strings.Builder
	for _, char := range msg {
		if slices.Contains(RUNES, char) {
			result.WriteRune(char)
		}
	}
	return result.String()
}
