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

func Generate(opts Options) (string, error) {
	if opts.Length < 8 {
		return "", fmt.Errorf("la longitud mínima es 8 caracteres")
	}
	charset := charsetBase
	if opts.Numbers {
		charset += charsetDigits
	}
	if opts.Symbols {
		charset += charsetSymbols
	}
	result := make([]byte, opts.Length)
	max := big.NewInt(int64(len(charset)))
	for i := range result {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("generator: generar caracter: %w", err)
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}
