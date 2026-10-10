package main

import (
	"fmt"
	"testing"
)

func passwordReader(answers ...string) func(string) (string, error) {
	i := 0
	return func(string) (string, error) {
		if i >= len(answers) {
			return "", fmt.Errorf("no more input")
		}
		answer := answers[i]
		i++
		return answer, nil
	}
}

func TestReadNewPassword(t *testing.T) {
	tests := []struct {
		name    string
		reader  func(string) (string, error)
		want    string
		wantErr bool
	}{
		{"valid", passwordReader("correct-horse", "correct-horse"), "correct-horse", false},
		{"empty", passwordReader("", ""), "", true},
		{"whitespace only", passwordReader("        ", "        "), "", true},
		{"too short", passwordReader("short", "short"), "", true},
		{"mismatch", passwordReader("correct-horse", "correct-horsf"), "", true},
		{"read error (no terminal)", func(string) (string, error) { return "", fmt.Errorf("inappropriate ioctl for device") }, "", true},
		{"read error on confirmation", passwordReader("correct-horse"), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readNewPassword(tt.reader)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
