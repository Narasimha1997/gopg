package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func sendError(w *http.ResponseWriter, message string) {
	data, _ := json.Marshal(MakeError(message))
	(*w).WriteHeader(http.StatusInternalServerError)
	(*w).Header().Set("Content-Type", "application/json")
	fmt.Fprintf(*w, string(data))
}

func sendInvalidMethod(w *http.ResponseWriter, message string) {
	data, _ := json.Marshal(MakeError(message))
	(*w).WriteHeader(http.StatusMethodNotAllowed)
	(*w).Header().Set("Content-Type", "application/json")
	fmt.Fprintf(*w, string(data))
}

func executeJSON(w *http.ResponseWriter, r *http.Request, channel chan<- bool) {
	contentType := r.Header.Get("Content-Type")

	spl := strings.Split(contentType, ";")
	if len(spl) > 0 {
		contentType = spl[0]
	}

	if r.Method != "POST" {
		sendInvalidMethod(w, fmt.Sprintf("Method %s not allowed", r.Method))
		channel <- true
		return
	}

	if contentType != "application/json" {
		sendError(w, "Content-Type must be application/json")
		channel <- true
		return
	}

	//execute the code
	input := InputPack{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendError(w, "Failed to parse body")
		channel <- true
		return
	}

	err = json.Unmarshal(body, &input)
	if err != nil {
		sendError(w, "Invalid json structure provided")
		channel <- true
		return
	}

	//execute the program
	programOutput := ExecuteTask(&input)
	if programOutput.Error {
		sendError(w, programOutput.ErrorString)
		channel <- true
		return
	}

	//success message
	bytes, err := json.Marshal(&programOutput)
	if err != nil {
		sendError(w, "Failed to serialize program output")
		channel <- true
		return
	}

	//send output
	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(http.StatusOK)
	fmt.Fprintf(*w, string(bytes))

	channel <- true
	return
}

func executeFile(w *http.ResponseWriter, r *http.Request, channel chan<- bool) {
	contentType := r.Header.Get("Content-Type")

	spl := strings.Split(contentType, ";")
	if len(spl) > 0 {
		contentType = spl[0]
	}

	if r.Method != "POST" {
		sendInvalidMethod(w, fmt.Sprintf("Method %s not allowed", r.Method))
		channel <- true
		return
	}

	if contentType != "multipart/form-data" {
		sendError(w, "Content-Type must be multipart/form-data")
		channel <- true
		return
	}

	//execute the code
	file, _, err := r.FormFile("file")
	if err != nil {
		sendError(w, "File not found, form data requires file attribute to be set")
		channel <- true
		return
	}

	defer file.Close()

	var buffer bytes.Buffer
	io.Copy(&buffer, file)

	program := buffer.String()
	input := InputPack{}
	input.Program = program

	//execute the program
	programOutput := ExecuteTask(&input)
	if programOutput.Error {
		sendError(w, programOutput.ErrorString)
		channel <- true
		return
	}

	//success message
	bytes, err := json.Marshal(&programOutput)
	if err != nil {
		sendError(w, "Failed to serialize program output")
		channel <- true
		return
	}

	//send output
	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(http.StatusOK)
	fmt.Fprintf(*w, string(bytes))

	buffer.Reset()

	channel <- true
	return
}

func main() {

	pool := NewRouteHandler(100, 100)
	pool.RegisterRoute("/executeJson", executeJSON)
	pool.RegisterRoute("/executeFile", executeFile)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pool.Dispatch(w, r)
	})

	http.ListenAndServe(":9000", nil)
}
