package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

const (
	colorReset  string = "\033[0m"
	colorRed    string = "\033[31m"
	colorGreen  string = "\033[32m"
	colorYellow string = "\033[33m"
	colorBlue   string = "\033[34m"
	colorPurple string = "\033[35m"
	colorCyan   string = "\033[36m"
	colorWhite  string = "\033[37m"
)

//ProgramOutput Represents the output of the program
type ProgramOutput struct {
	Success       bool    `json:"success"`
	Output        string  `json:"output"`
	ExecutionTime float64 `json:"executionTime"`
}

//OutputPack Represents the output package
type OutputPack struct {
	Error       bool          `json:"error"`
	ErrorString string        `json:"errorString"`
	Output      ProgramOutput `json:"execution"`
}

func makeRequest(filename string) *OutputPack {
	//check if file exist
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		log.Fatalf("File %s does not exist", filename)
	}

	if info.Mode().IsDir() {
		log.Fatalf("The path %s is a directory", filename)
	}

	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		log.Fatalf("Failed to open file %s\n", filename)
	}

	var buffer bytes.Buffer
	defer buffer.Reset()
	var fileWriter io.Writer

	writer := multipart.NewWriter(&buffer)

	fileWriter, err = writer.CreateFormFile("file", filename)
	if err != nil {
		log.Fatalf("Failed to read file %s\n", filename)
	}

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		log.Fatalf("Failed to read file %s\n", filename)
	}

	writer.Close()

	//make post request
	uri, exists := os.LookupEnv("GOPG_URL")
	if !exists {
		uri = "http://localhost:9000"
	}

	if strings.HasSuffix(uri, "/") {
		uri = uri[:len(uri)-1]
	}

	uri = uri + "/executeFile"
	fmt.Println(uri)
	request, err := http.NewRequest("POST", uri, &buffer)
	if err != nil {
		log.Fatalf("Failed to create request\n")
	}

	request.Header.Set("Content-type", writer.FormDataContentType())
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatalf("Failed to make request\n")
	}

	defer response.Body.Close()

	//parse the response
	body, err := ioutil.ReadAll(response.Body)

	programOutput := OutputPack{}
	err = json.Unmarshal(body, &programOutput)
	if err != nil {
		log.Fatalf("Failed to parse output\n")
	}

	//return the output
	return &programOutput
}

func pprint(response *OutputPack) {
	fmt.Println("Server Status:")
	fmt.Println("-------------------")
	if response.Error {
		fmt.Println(string(colorRed), "Server encountered an error processing the file")
		fmt.Println(string(colorRed), "Error message: ", response.ErrorString)
		return
	}

	fmt.Println(string(colorGreen), "Server successfully processed your file")
	fmt.Println(string(colorReset))
	fmt.Println("============================================================")
	fmt.Println("Program output:")
	fmt.Println("-------------------")

	if response.Output.Success {
		fmt.Println(string(colorGreen), response.Output.Output)
		fmt.Println(string(colorReset))
		fmt.Println("============================================================")
		fmt.Println("Compile + Execution time:")
		fmt.Println("-------------------", string(colorYellow))
		fmt.Printf("%f\n", response.Output.ExecutionTime)
		fmt.Println(string(colorReset))
		return
	}
	fmt.Println(string(colorRed), "Error")
	fmt.Println(string(colorRed), response.Output.Output)
	fmt.Println(string(colorReset))
	fmt.Println(string(colorReset))
	fmt.Println("============================================================")
	fmt.Println("Compile + Execution time:")
	fmt.Println("-------------------", string(colorYellow))
	fmt.Printf("%f\n", response.Output.ExecutionTime)

	fmt.Println(string(colorReset))
}

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatal("File path must be provided as an argument\n")
	}

	response := makeRequest(args[1])
	pprint(response)

	os.Exit(0)
}
