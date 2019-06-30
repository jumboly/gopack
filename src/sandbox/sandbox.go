package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const root = `d:\_\gopack\testdata\00000`

func main() {
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		fmt.Printf("path: %s, isDir: %t\n", path, info.IsDir())
		return nil
	})
}
