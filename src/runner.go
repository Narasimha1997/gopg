package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
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

//InputPack Represents input package
type InputPack struct {
	Program string `json:"program"`
}

//GoRunner compiles and runs a go-program
type GoRunner struct {
	programOutput *ProgramOutput
}

//B64Mapping mapping of base63 values
const B64Mapping string = "abcdefghijklmnopqrstuvwxyz"

func (g *GoRunner) generateRandonName() (string, error) {
	rHolder := make([]byte, 8)
	_, err := rand.Read(rHolder)

	if err != nil {
		log.Fatal("Failed to generate randon uuid string")
		return "", err
	}

	rInt := binary.BigEndian.Uint64(rHolder)

	mappedString := make([]byte, 0)

	//map to alphabets range
	for rInt > 0 {
		rem := rInt % 26
		rInt = rInt / 26
		mappedString = append(mappedString, B64Mapping[rem])
	}

	return string(mappedString[:]), nil
}

func (g *GoRunner) cleanUp(fp *string) error {
	err := os.Remove(*fp)
	return err
}

func (g *GoRunner) onResult(output *string, timeDiff *float64, err error, success bool) (*ProgramOutput, error) {
	programOutput := &ProgramOutput{
		Success:       success,
		ExecutionTime: *timeDiff,
		Output:        *output,
	}

	return programOutput, err
}

func (g *GoRunner) executeTask(goProgram *[]byte) (*ProgramOutput, error) {
	b63, err := g.generateRandonName()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	b63GoFile := "/tmp/" + b63 + ".go"

	//save output to /tmp
	err = ioutil.WriteFile(b63GoFile, *goProgram, 0666)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//execute the go-code with stderr and stdout connectors
	executor := exec.Command("timeout", "10", "go", "run", b63GoFile)

	st := time.Now()
	output, err := executor.CombinedOutput()
	et := time.Now()

	outputStr := string(output)

	tdiff := et.Sub(st).Seconds()

	if tdiff >= 10 {
		errString := "Execution timeout"
		err = errors.New("Execution timeout error")
		g.cleanUp(&b63GoFile)
		return g.onResult(&errString, &tdiff, err, false)
	}

	if err != nil {
		g.cleanUp(&b63GoFile)
		return g.onResult(&outputStr, &tdiff, err, false)
	}

	g.cleanUp(&b63GoFile)
	return g.onResult(&outputStr, &tdiff, err, true)
}

//MakeError Returns an  error object
func MakeError(errString string) *OutputPack {
	return &OutputPack{
		Error:       true,
		ErrorString: errString,
		Output: ProgramOutput{
			Success:       false,
			Output:        "",
			ExecutionTime: 0.00,
		},
	}
}

//ExecuteTask Executes a program
func ExecuteTask(inputPack *InputPack) *OutputPack {
	if inputPack.Program == "" || inputPack.Program == " " {
		return &OutputPack{
			Error:       true,
			ErrorString: "Empty program found",
			Output: ProgramOutput{
				Success:       false,
				Output:        "",
				ExecutionTime: 0.000,
			},
		}
	}

	executor := GoRunner{}

	pBytes := []byte(inputPack.Program)
	programOutput, err := executor.executeTask(&pBytes)

	if err != nil && programOutput == nil {
		return &OutputPack{
			Error:       true,
			ErrorString: "General execution error",
			Output: ProgramOutput{
				Success:       false,
				Output:        "",
				ExecutionTime: 0.000,
			},
		}
	}

	return &OutputPack{
		Error:       false,
		ErrorString: "",
		Output: ProgramOutput{
			Success:       programOutput.Success,
			Output:        programOutput.Output,
			ExecutionTime: programOutput.ExecutionTime,
		},
	}
}
