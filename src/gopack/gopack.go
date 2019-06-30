package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"./commons"
)

// eachDir dir を指定された level 階層リストする
func eachDir(dir string, level int, action func(string, os.FileInfo) error) error {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, info := range infos {
		path := filepath.Join(dir, info.Name())

		if !info.IsDir() {
			if err := action(path, info); err != nil {
				return err
			}
		} else {
			if level == 0 {
				if err := action(path, info); err != nil {
					return err
				}
			} else {
				if err := eachDir(path, level-1, action); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func writeTar(tw *tar.Writer, dir string, name string) error {
	path := filepath.Join(dir, name)

	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("ディレクトリは対象外")
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = name

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	fr, err := os.Open(path)
	if err != nil {
		return err
	}

	if _, err := io.Copy(tw, fr); err != nil {
		return err
	}

	return nil
}

func pack(inDir, outFile string) error {

	fw, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer fw.Close()

	gw := gzip.NewWriter(fw)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	err = filepath.Walk(inDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		log.Printf("pack: %s", path)

		path = strings.TrimPrefix(path, inDir)
		return writeTar(tw, inDir, path)
	})
	if err != nil {
		return err
	}

	return nil
}

func main() {
	level := flag.Int("l", 1, "圧縮する階層")
	flag.Parse()

	inRoot := flag.Arg(0)  // `d:\_\gopack\testdata2\`
	outRoot := flag.Arg(1) //`d:\_\gopack\pack2\`

	inRoot = commons.PathClearn(inRoot)
	outRoot = commons.PathClearn(outRoot)

	err := eachDir(inRoot, *level, func(path string, info os.FileInfo) error {
		outPath := filepath.Join(outRoot, strings.TrimPrefix(path, inRoot))

		if err := os.MkdirAll(filepath.Dir(outPath), commons.DirPerm); err != nil {
			return err
		}

		if info.IsDir() {
			tarName := outPath + commons.PacExt
			log.Printf("tar: %s -> %s\n", path, tarName)
			if err := pack(path, tarName); err != nil {
				return err
			}
		} else {
			log.Printf("fileCopy: %s -> %s\n", path, outPath)
			if err := commons.FileCopy(path, outPath); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}
