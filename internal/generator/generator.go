package generator

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	charsetBase    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetDigits  = "0123456789"
	charsetSymbols = "@#$%&*()!"
)

type Options struct {
	Length  int
	Numbers bool
	Symbols bool
}

// Generate returns a random password that is guaranteed to contain at least
// one character from every enabled class (letters, digits, symbols).
func Generate(opts Options) (string, error) {
	if opts.Length < 8 {
		return "", fmt.Errorf("la longitud mínima es 8 caracteres")
	}

	classes := []string{charsetBase}
	if opts.Numbers {
		classes = append(classes, charsetDigits)
	}
	if opts.Symbols {
		classes = append(classes, charsetSymbols)
	}

	charset := ""
	for _, c := range classes {
		charset += c
	}

	result := make([]byte, opts.Length)
	// One guaranteed character per enabled class, the rest from the full charset.
	for i := range result {
		pool := charset
		if i < len(classes) {
			pool = classes[i]
		}
		n, err := randInt(len(pool))
		if err != nil {
			return "", err
		}
		result[i] = pool[n]
	}

	// Fisher-Yates shuffle so guaranteed characters are not at fixed positions.
	for i := len(result) - 1; i > 0; i-- {
		j, err := randInt(i + 1)
		if err != nil {
			return "", err
		}
		result[i], result[j] = result[j], result[i]
	}
	return string(result), nil
}

func randInt(max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, fmt.Errorf("generator: generar caracter: %w", err)
	}
	return int(n.Int64()), nil
}
