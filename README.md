### gopg
A minimal microservice written in Go that executes Go programs. This microservice can be used to set-up local go learning environment at your workspace/school. You can also use the provided zero-configuration docker-image for quick deployments.

#### Features:
1. Zero-dependencies.
2. Simple POST APIs to execute from JSON or File.
3. A built-in work-queue that takes care of parallelizing execution work-loads.
4. Auto-suspension of executions taking more than 10 seconds - avoids infinite looping.
5. Easily deployable using docker image.
6. Read-only file-system for execution instance (TODO)

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

#### Docker set-up
To build the docker image, you can run:
```
docker build . -t gopg:latest
```

Then, run the image using:
```
docker run --rm -ti -p 9000:9000 gopg:latest 
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

#### Contributing
Contributions are always welcome. You can raise an issue or contribute new features by making a PR.