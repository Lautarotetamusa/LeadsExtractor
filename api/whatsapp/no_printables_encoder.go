package whatsapp

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
)

var RUNES = []rune{'\u200a', '\u200b', '\u200c', '\u200d', '\u200e'}

const RUNE_LEN = 3

func encodeString(s string) string {
	var encoded strings.Builder
	for _, r := range s {
		encodedChar := encode(int(r))
        // Restamos uno porque len(encodedChar) nunca será menor a 1. 
        // entonces encodeamos len=1 como len=0. Esto nos da 5 veces mas combinaciones
        length := (len(encodedChar)-1) / RUNE_LEN

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
        if (i+RUNE_LEN > len(s)){
            return "", fmt.Errorf("decoded string out of index")
        }

		lengthStr := s[i : i+RUNE_LEN]
		length := decode(lengthStr)
		if length == -1 {
            return "", fmt.Errorf("error decoding runes %q", lengthStr)
		}
		i += RUNE_LEN

        strLen := (length+1)*RUNE_LEN
        if (i+strLen > len(s)){
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

func GenerateWppLink(phone string, params string, msg string) string {
    const baseUrl = "https://wa.me/%s?text=%s%s"
	return fmt.Sprintf(baseUrl, phone, url.QueryEscape(encodeString(params)), msg)
}

func main() {
    fmt.Printf("encode(110)=%q\n", encode(110))
    fmt.Printf("encode(4)=%q\n", encode(4)) // hasta 24 lo encodeas con 2 caracteres
    fmt.Printf("encode(24)=%q\n", encode(24)) // hasta 24 lo encodeas con 2 caracteres
    fmt.Printf("encode(26)=%q\n", encode(25)) 
    fmt.Printf("encode(124)=%q\n", encode(124)) // hasta 124 lo encodeas con 3 caracteres
    fmt.Printf("encode(125)=%q\n", encode(125)) 
    // 4 runes -> encodeas 5**4-1 = 625-1 combinaciones (alcanza para ascci)

    params := "s=rebora&m=ashe&c=test"
	encoded := encodeString(params)
	fmt.Printf("Encoded: %q\n", encoded)
	decoded, err := decodeString(encoded)
    if err != nil {
        fmt.Printf("err=%s\n", err.Error())
    }
	fmt.Println("Decoded:", decoded)
    _, err = decodeString("\u200e\u200e\u200c\u200c")
    fmt.Println("Error:", err.Error())
    fmt.Println(GenerateWppLink("5213328092850", params, "Hola!"));
}

/*
cedbcebcedcedbHola!
cedbcebcedcedb
cebacccbcbeecbdececacccbccab

d -> b
b -> a
e -> c 
b -> c |

https://wa.me/5213328092850?text=hola%5Cu200b%5Cu200a%5Cu200ahola

\u200c\u200c\u200e\u200d\u200b\u200c\u200e\u200a\u200bHola!

\u200c\u200e\u200d\u200b\u200c\u200e\u200b
\u200c\u200e\u200d\u200b\u200c\u200e\u200a\u200b
%E2%80%8C%E2%80%8E%E2%80%8D%E2%80%8B%E2%80%8C%E2%80%8E%E2%80%8A%E2%80%8BHola%21

\u200c\u200e\u200d\u200b\u200c\u200e \u200bHola
\u200c\u200e\u200d\u200b\u200c\u200e\u200b



\u200c\u200e\u200d\u200b\u200c\u200e\u200a\u200b
\u200c\u200e\u200d\u200b\u200c\u200e      \u200b
\u200c\u200c\u200e\u200d\u200b\u200c\u200e\u200a\u200b
*/
