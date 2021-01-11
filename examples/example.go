package main

import (
	"fmt"
	"io/ioutil"
	"runtime"
)

func main() {
	fmt.Println("Hello, world!!")
	fmt.Println(runtime.GOOS)

	files, _ := ioutil.ReadDir("/")

	fmt.Println("Files :")

	for _, file := range files {
		fmt.Println(file.Name())
	}
}
