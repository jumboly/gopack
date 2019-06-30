package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"./commons"
)

type File struct {
	Name  string
	Bytes []byte
}

func readFile(tr *tar.Reader, parallels int) <-chan interface{} {
	ch := make(chan interface{}, parallels)
	go func() {
		defer close(ch)

		for {
			header, err := tr.Next()
			if err != nil {
				ch <- err
				break
			}

			bytes, err := ioutil.ReadAll(tr)
			if err != nil {
				ch <- err
				break
			}

			ch <- File{header.Name, bytes}
		}
	}()
	return ch
}

func unpack(inFile, outDir string, parallels int) error {
	fr, err := os.Open(inFile)
	if err != nil {
		return err
	}
	defer fr.Close()

	gr, err := gzip.NewReader(fr)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	ch := readFile(tr, parallels)

	wait := new(sync.WaitGroup)
	for data := range ch {

		switch d := data.(type) {
		case File:
			wait.Add(1)

			go func() {
				defer wait.Done()

				outFile := filepath.Join(outDir, d.Name)
				if err := os.MkdirAll(filepath.Dir(outFile), commons.DirPerm); err != nil {
					log.Fatal(err)
				}
				if err := ioutil.WriteFile(outFile, d.Bytes, commons.FilePerm); err != nil {
					log.Fatal(err)
				}
			}()

		case error:
			if d != io.EOF {
				return d
			}
		}
	}

	wait.Wait()

	return nil
}

func untar(tarFile, outDir string) <-chan File {
	ch := make(chan File)
	go func() {
		defer close(ch)

		fr, err := os.Open(tarFile)
		commons.IfErrorFatal(err)
		defer fr.Close()

		gr, err := gzip.NewReader(fr)
		commons.IfErrorFatal(err)
		defer gr.Close()

		tr := tar.NewReader(gr)

		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			commons.IfErrorFatal(err)

			bytes, err := ioutil.ReadAll(tr)
			commons.IfErrorFatal(err)

			ch <- File{
				Name:  filepath.Join(outDir, header.Name),
				Bytes: bytes,
			}
		}
	}()
	return ch
}

func getFile(inRoot, outRoot string, parallels int) <-chan File {
	ch := make(chan File, parallels)
	go func() {
		defer close(ch)

		//lim := make(chan int, parallels*2)
		_ = filepath.Walk(inRoot, func(path string, info os.FileInfo, err error) error {
			commons.IfErrorFatal(err)

			if info.IsDir() {
				return nil
			}

			outPath := filepath.Join(outRoot, strings.TrimPrefix(path, inRoot))

			ext := filepath.Ext(path)
			if ext == commons.PacExt {
				outDir := strings.TrimSuffix(outPath, ext)
				log.Printf("unpack: %s -> %s", path, outDir)
				for file := range untar(path, outDir) {
					ch <- file
				}
			} else {
				log.Printf("copy: %s -> %s", path, outPath)

				bytes, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}

				ch <- File{outPath, bytes}
			}

			return nil
		})
	}()
	return ch
}

func main() {
	parallels := flag.Int("p", runtime.NumCPU(), "並列数")
	flag.Parse()

	inRoot := flag.Arg(0)  //`d:\_\gopack\pack`
	outRoot := flag.Arg(1) //`d:\_\gopack\unpack`

	start := time.Now()

	inRoot = commons.PathClearn(inRoot)
	outRoot = commons.PathClearn(outRoot)

	log.Printf("GOMAXPROCS: %d", runtime.GOMAXPROCS())
	runtime.GOMAXPROCS(*parallels)

	wait := new(sync.WaitGroup)

	//lim := make(chan struct{}, parallels*2)
	//for file := range getFile(inRoot, outRoot, parallels) {
	//
	//	dir := filepath.Dir(file.Name)
	//	if _, ok := dirs[dir]; !ok {
	//		dirs[dir] = struct{}{}
	//		if err := os.MkdirAll(filepath.Dir(file.Name), commons.DirPerm); err != nil {
	//			commons.IfErrorFatal(err)
	//		}
	//	}
	//
	//	select {
	//	case lim <- struct{}{}:
	//		wait.Add(1)
	//
	//		go func(file File) {
	//			defer func() {
	//				<-lim
	//				wait.Done()
	//			}()
	//
	//			if err := ioutil.WriteFile(file.Name, file.Bytes, commons.FilePerm); err != nil {
	//				commons.IfErrorFatal(err)
	//			}
	//		}(file)
	//	}
	//}

	dirs := make(map[string]struct{})
	for file := range getFile(inRoot, outRoot, parallels) {

		dir := filepath.Dir(file.Name)
		if _, ok := dirs[dir]; !ok {
			dirs[dir] = struct{}{}
			if err := os.MkdirAll(dir, commons.DirPerm); err != nil {
				commons.IfErrorFatal(err)
			}
		}

		//if err := os.MkdirAll(filepath.Dir(file.Name), commons.DirPerm); err != nil {
		//	commons.IfErrorFatal(err)
		//}

		wait.Add(1)

		go func(file File) {
			defer wait.Done()

			if err := ioutil.WriteFile(file.Name, file.Bytes, commons.FilePerm); err != nil {
				commons.IfErrorFatal(err)
			}
		}(file)
	}

	wait.Wait()

	//err := filepath.Walk(inRoot, func(path string, info os.FileInfo, err error) error {
	//	if err != nil {
	//		return err
	//	}
	//
	//	if info.IsDir() {
	//		return nil
	//	}
	//
	//	outPath := filepath.Join(outRoot, strings.TrimPrefix(path, inRoot))
	//
	//	if err := os.MkdirAll(filepath.Dir(outPath), commons.DirPerm); err != nil {
	//		return err
	//	}
	//
	//	ext := filepath.Ext(path)
	//	if ext == commons.PacExt {
	//		outDir := strings.TrimSuffix(outPath, ext)
	//		log.Printf("unpack: %s -> %s", path, outDir)
	//		if err := os.MkdirAll(outDir, commons.DirPerm); err != nil {
	//			return err
	//		}
	//		if err := unpack(path, outDir, parallels); err != nil {
	//			return err
	//		}
	//	} else {
	//		log.Printf("copy: %s -> %s", path, outPath)
	//		if err := commons.FileCopy(path, outPath); err != nil {
	//			return err
	//		}
	//	}
	//
	//	return nil
	//})
	//
	//if err != nil {
	//	log.Fatal(err)
	//}

	log.Printf("処理時間: %f", time.Now().Sub(start).Seconds())
}
