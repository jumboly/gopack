package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
)

// randMinMax min ～ max の乱数を生成
func randMinMax(min, max int) int {
	return min + rand.Intn(max-min)
}

// exists 指定した path が存在するか.
func exists(path string) bool {
	// path が存在しない場合はエラー
	_, err := os.Stat(path)
	return err == nil
}

func createLevel1(parent string) {
	for i := 0; i < 10; i++ {
		dirname := path.Join(parent, fmt.Sprintf("%05d", i))
		if err := os.Mkdir(dirname, 0777); err != nil {
			log.Fatal(err)
		}
		log.Printf("Mkdir:%s", dirname)
		createLevel2(dirname)
	}

	for i := 0; i < 5; i++ {
		filename := path.Join(parent, fmt.Sprintf("%05d.dat", i))
		createDummyFile(filename, randMinMax(1000, 50000))
	}
}

func createLevel2(parent string) {
	for i := 0; i < randMinMax(10, 20); i++ {
		dirname := path.Join(parent, fmt.Sprintf("%05d", i))
		if err := os.Mkdir(dirname, 0777); err != nil {
			log.Fatal(err)
		}
		log.Printf("Mkdir:%s", dirname)
		createLevel3(dirname)
	}

	for i := 0; i < randMinMax(10, 20); i++ {
		filename := path.Join(parent, fmt.Sprintf("%05d.dat", i))
		createDummyFile(filename, randMinMax(1000, 50000))
	}
}

func createLevel3(parent string) {
	for i := 0; i < randMinMax(10, 30); i++ {
		dirname := path.Join(parent, fmt.Sprintf("%05d", i))
		if err := os.Mkdir(dirname, 0777); err != nil {
			log.Fatal(err)
		}
		log.Printf("Mkdir:%s", dirname)
		createLevel4(dirname)
	}

	for i := 0; i < randMinMax(10, 30); i++ {
		filename := path.Join(parent, fmt.Sprintf("%05d.dat", i))
		createDummyFile(filename, randMinMax(1000, 50000))
	}
}

func createLevel4(parent string) {
	for i := 0; i < randMinMax(10, 100); i++ {
		filename := path.Join(parent, fmt.Sprintf("%05d.dat", i))
		createDummyFile(filename, randMinMax(1000, 50000))
	}
}

func createDummyFile(filename string, size int) {
	bytes := make([]byte, size)
	rand.Read(bytes)
	if err := ioutil.WriteFile(filename, bytes, 0777); err != nil {
		log.Fatal(err)
	}
}

func main() {
	rand.Seed(1)

	root := os.Args[1]
	clear(root)
	createLevel1(root)
}
