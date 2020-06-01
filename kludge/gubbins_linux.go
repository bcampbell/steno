package kludge

import (
	"fmt"
	"os"
	"path/filepath"
)

func DataPath() (string, error) {
	return ".", nil
}

// get (or create) per-user directory (eg "$HOME/.steno")
func PerUserPath() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("$HOME not set")
	}
	dir := filepath.Join(home, ".steno")
	// create dir if if doesn't already exist
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}
	return dir, nil
}

// path to any external tool binaries (eg fasttext)
func BinPath() (string, error) {
	datPath, err := DataPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(datPath, "bin"), nil
}
