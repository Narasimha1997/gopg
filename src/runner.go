package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const (

	//SandboxMemory main memory to be used - 500MB of memory
	SandboxMemory int = (1 << 20) * 500

	//SandboxTimeout timeout of the sandbox in seconds
	SandboxTimeout = 10
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

func timerStart(channel chan<- bool) {
	time.Sleep(SandboxTimeout * time.Second)
	channel <- true
}

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

func (g *GoRunner) isSandboxEnabled() bool {
	value, exist := os.LookupEnv("SANDBOX")

	fmt.Println(value)
	if !exist {
		return false
	}
	if value == "1" {
		return true
	}
	return false
}

func (g *GoRunner) sandboxExecute(goFile string) (*ProgramOutput, error) {

	goFileSource := goFile
	goFile = strings.ReplaceAll(goFile, ".", "_")

	fmt.Println("Binary : " + goFile)

	command := fmt.Sprintf("go build -ldflags '-w -extldflags \"-static\"' -o %s -i %s\n", goFile, goFileSource)

	fmt.Printf("Command %s\n", command)

	compiler := exec.Command("/bin/bash", "-c", command)

	tstart := time.Now()
	err := compiler.Run()
	tend := time.Now()

	compileTime := tend.Sub(tstart).Seconds()

	if err != nil {
		output, err := compiler.CombinedOutput()
		outputString := string(output)
		g.cleanUp(&goFileSource)
		g.onResult(&outputString, &compileTime, err, false)
	}

	containerName := strings.ReplaceAll(goFile, "/", "")

	//compilation is successful, not start the container and pass stdin
	executor := exec.Command(
		"docker", "run",
		"--runtime=runsc",
		"--memory="+fmt.Sprintf("%d", SandboxMemory),
		"--name="+containerName,
		"-i",
		"sandbox:latest",
	)
	executor.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var outputBuffer strings.Builder

	data, err := ioutil.ReadFile(goFile)
	if err != nil {
		outputString := "Failed to open binary file for reading"
		g.cleanUp(&goFileSource)
		g.cleanUp(&goFile)
		return g.onResult(&outputString, &compileTime, err, false)
	}

	//attach the stdin and write data to child
	stdin, err := executor.StdinPipe()
	if err != nil {
		outputString := "Failed to get stdin of the child-process"
		return g.onResult(&outputString, &compileTime, err, false)
	}

	stdout, err := executor.StdoutPipe()

	if err != nil {
		fmt.Println(err)
		outputString := "Failed to run command"
		g.cleanUp(&goFileSource)
		g.cleanUp(&goFile)
		return g.onResult(&outputString, &compileTime, err, false)
	}

	stderr, err := executor.StderrPipe()

	if err != nil {
		fmt.Println(err)
		outputString := "Failed to run command"
		g.cleanUp(&goFileSource)
		g.cleanUp(&goFile)
		return g.onResult(&outputString, &compileTime, err, false)
	}

	//We will use this function to clean the child process
	childProcessCleaner := func(exitProcess bool) bool {

		if exitProcess {
			pgid, err := syscall.Getpgid(executor.Process.Pid)
			if err != nil {
				return false
			}
			syscall.Kill(-pgid, syscall.SIGTERM)
		}

		stdout.Close()
		stderr.Close()
		g.cleanUp(&goFileSource)
		g.cleanUp(&goFile)

		return true
	}

	executionEnd := make(chan error, 1)

	processWaiter := func() {
		err := executor.Wait()
		executionEnd <- err
	}

	err = executor.Start()
	go processWaiter()

	stdin.Write(data)

	//closes the stdin channel properly with EOF
	stdin.Close()
	executionStartTime := time.Now()

	go io.Copy(&outputBuffer, stdout)
	go io.Copy(&outputBuffer, stderr)

	if err != nil {
		executionEndTime := time.Now()
		totalTime := executionEndTime.Sub(executionStartTime).Seconds() + compileTime
		fmt.Println(err)
		outputString := "Failed to execute the container, internal error"
		childProcessCleaner(false)
		return g.onResult(&outputString, &totalTime, err, false)
	}

	select {
	case err := <-executionEnd:
		if err != nil {
			executionEndTime := time.Now()
			totalTime := executionEndTime.Sub(executionStartTime).Seconds() + compileTime
			outputMessage := "Wait error"
			childProcessCleaner(false)
			return g.onResult(&outputMessage, &totalTime, err, true)
		}

		//executed gracefully
		executionEndTime := time.Now()
		totalTime := executionEndTime.Sub(executionStartTime).Seconds() + compileTime
		childProcessCleaner(false)
		result := outputBuffer.String()
		return g.onResult(&result, &totalTime, nil, true)
	case <-time.After(SandboxTimeout * time.Second):
		executionEndTime := time.Now()
		totalTime := executionEndTime.Sub(executionStartTime).Seconds() + compileTime
		//timeout error
		outputString := outputBuffer.String() + "\n[Execution Timeout]\n"
		//terminate and exit
		childProcessCleaner(true)
		return g.onResult(&outputString, &totalTime, nil, false)
	}
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

	if g.isSandboxEnabled() {
		fmt.Println("Sandbox enabled, running in sandbox")
		return g.sandboxExecute(b63GoFile)
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
