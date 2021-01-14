### gopg
A minimal microservice written in Go that executes Go programs. This microservice can be used to set-up local go learning environment at your workspace/school. The tool comes with an optional secure sandbox powered by gVisor and Docker.

#### Features:
1. Zero-dependencies.
2. Simple POST APIs to execute from JSON or File.
3. A built-in work-queue that takes care of parallelizing execution work-loads.
4. Auto-suspension of executions taking more than 10 seconds - avoids infinite looping.
5. Easily deployable using docker image.
6. Secure sandbox which executes the programs in an isolated environment. 

#### Local set-up
To install this locally, clone the repository and build the binary by running:
```
./build.sh
```

Run the binary:
```
./bin/gopg
```

This requires Go programming language to be installed and working correctly on the local machine.


### Using gopg
There are two ways you can use `gopg`:

Requirements:
1. Golang 1.5+
2. GCC compiler
3. Docker installed and configured.
4. `runsc` - gVisor runtime pluin for docker, you can install it by running `scripts/install_runsc.sh`

#### 1. Local Setup
To build and use gopg locally, go to `./scripts` and run:
```
cd scripts/
./build_sandbox.sh
```
Then you can run `gopg` as :
```
./bin/gopg
```

#### 2. Docker-setup
To run the entire environment inside a docker container, run the same command with `--docker` option.
```
cd scripts/
./build_sandbox.sh --docker
```

Then you can start the container as follows: (you need to pass docker socket file to the container)
```
docker run -ti -v /var/run/docker.sock:/var/run/docker.sock --net=host gopg
```

#### Enabling Sandboxed mode
The sandbox mode can be enabled/disabled whenever required. (Note : Running without sandbox can execute the binary directly on your host kernel and has access to the host-file system which is not recommended). In some scenarios, you may need not have to worry about security, in such cases you can turn off the sandbox. If you need all the security features to be available, you can enable sandbox (Note : Sandboxed mode introduces more latency because the container needs to be created with gVisor runtime everytime you execute the program). 

To enable sandbox, you can set `SANDBOX=1` environment variable, `gopg` sees this environment variable to decide whether to run sandbox or not. 

Locally:
```
export SANDBOX=1
./bin/gopg
```

Docker:
```
docker run -ti -v /var/run/docker.sock:/var/run/docker.sock --net=host --env="SANDBOX=1" gopg
```

#### Example API usage
The API `/executeJSON` can be used to execute go-programs. Let's create a simple json structure like the one shown below (example.json):

```json
{
    "program" : "package main\n\nimport \"fmt\"\n\nfunc main() {\n fmt.Println(\"Hello, world!\")\n\n }"
}
```

And use curl to send a json request
```
curl -X POST -H "Content-Type: application/json" -d @./examples/example.json http://localhost:9000/executeJson | json_pp
```

You should see the following output:
```json
{
   "error" : false,
   "errorString" : "",
   "execution" : {
      "output" : "Hello, world!\n",
      "executionTime" : 0.196068332,
      "success" : true
   }
}
```
The keys `error` and `errorString` will contain API errors, the output information can be found inside `execution` key. The `output` contains output string or the error string in case of runtime/syntax errors. The `executionTime` key says the execution time in seconds and finally the `success` key will say if the program executed successfully or encountered an error.

You can also use the File API which takes `multipart/form-data` as input and provides the result. Let's create a file called `example.go` under `examples`:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, world!!")
}

```

Now we can send `multipart/form-data` to the server, using `curl`

```
curl -F file=@./examples/example.go -H "Content-Type:multipart/form-data" http://localhost:9000/executeFile | json_pp
```

You can see the output exactly like the previous case:
```json
{
   "execution" : {
      "output" : "Hello, world!!\n",
      "executionTime" : 0.220491135,
      "success" : true
   },
   "error" : false,
   "errorString" : ""
}
```

#### Using client-binary
`./script/build_client.sh` builds client binary. The client binary executes go-programs by making request to the server. You can use the client binary as follows:

```
./bin/gopg-client ./examples/example.go
```

If everything worked as expected, it should produce the output as shown below:

```
Server Status:
-------------------
 Server successfully processed your file

============================================================
Program output:
-------------------
 Hello, world!!


============================================================
Compile + Execution time:
------------------- 
0.554610
```

#### Contributing
Contributions are always welcome. You can raise an issue or contribute new features by making a PR.