package commons

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	// PacExt アーカイブファイルの拡張子
	PacExt = ".gopack"
	// DirPerm ディレクトリのパーミッション
	DirPerm = 0777
	// FilePerm ファイルのパーミッション
	FilePerm = 0666
	// PathSeparator パスの区切り文字
	PathSeparator = string(filepath.Separator)
)

func IfErrorFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// PathClearn パスの区切り文字を統一する.
func PathClearn(path string) string {
	// windowsはパス区切りに \ と / が使えるので統一する

	path = strings.Map(func(r rune) rune {
		if os.IsPathSeparator(uint8(r)) {
			return filepath.Separator
		} else {
			return r
		}
	}, path)
	path = strings.TrimSuffix(path, PathSeparator)
	return path
}

// FileCopy src を dst にコピーする.
func FileCopy(src, dst string) error {
	sr, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sr.Close()

	dw, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dw.Close()

	_, err = io.Copy(dw, sr)
	if err != nil {
		return err
	}

	return nil
}
