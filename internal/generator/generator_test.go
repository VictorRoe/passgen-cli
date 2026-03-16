package generator_test

import (
	"strings"
	"testing"

	"github.com/VictorRoe/passgen-cli/internal/generator"
)

func TestGenerate_Length(t *testing.T) {
	opts := generator.Options{Length: 16}
	got, err := generator.Generate(opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(got) != 16 {
		t.Errorf("Generate() len = %d, want 16", len(got))
	}
}

func TestGenerate_LengthTooShort(t *testing.T) {
	opts := generator.Options{Length: 7}
	_, err := generator.Generate(opts)
	if err == nil {
		t.Error("Generate() expected error for length < 8, got nil")
	}
}

func TestGenerate_NoSymbolsByDefault(t *testing.T) {
	opts := generator.Options{Length: 200}
	got, err := generator.Generate(opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	const syms = "@#$%&*()!"
	for _, c := range got {
		if strings.ContainsRune(syms, c) {
			t.Errorf("Generate() without Symbols=true, found symbol %q", c)
		}
	}
}

func TestGenerate_NoNumbersByDefault(t *testing.T) {
	opts := generator.Options{Length: 200}
	got, err := generator.Generate(opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	const digits = "0123456789"
	for _, c := range got {
		if strings.ContainsRune(digits, c) {
			t.Errorf("Generate() without Numbers=true, found digit %q", c)
		}
	}
}

func TestGenerate_WithNumbers(t *testing.T) {
	opts := generator.Options{Length: 200, Numbers: true}
	got, err := generator.Generate(opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	const digits = "0123456789"
	hasDigit := false
	for _, c := range got {
		if strings.ContainsRune(digits, c) {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		t.Error("Generate() with Numbers=true: no digit found in 200-char password")
	}
}

func TestGenerate_WithSymbols(t *testing.T) {
	opts := generator.Options{Length: 200, Symbols: true}
	got, err := generator.Generate(opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	const syms = "@#$%&*()!"
	hasSym := false
	for _, c := range got {
		if strings.ContainsRune(syms, c) {
			hasSym = true
			break
		}
	}
	if !hasSym {
		t.Error("Generate() with Symbols=true: no symbol found in 200-char password")
	}
}
