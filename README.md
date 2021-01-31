# tini-and-distroless-poc
Demonstration of how you can use tini-static and the distroless base image for go

### Statically vs Dynamically Linking
A statically linked program has all of it's dependencies compiled into the executable. In this case it does not need to load or utilize any other libraries or code. While a dynamically linked program has external dependencies that need to be present and are loaded during execution.

One of the reasons I love working with GO is because I can compile go code into a statically linked binary. Why bother? Well, it's hard to answer but simply putting it everything is needs is right inside the binary as opposed to a dynamically linked binary. A dynamically linked binary is usually smaller in size as it dynamically links to it's dependencies installed on an OS. Lets take LibC for example, LibC contain certain library functions required by the application code, when I run a dynamically linked binary, it will link to whatever LibC found on the system whereas a statically linked binary has LibC (and other dependent libraries) embeded inside the binary itself hence statically linked binary is larger in size but does not use any of the system libraries. It has many advantages but those are out of scope of this article. This article answers How and assumes you know Why.

> Credit: https://oddcode.daveamit.com/2018/08/16/statically-compile-golang-binary/

### Container init process
A container’s main running process is the ENTRYPOINT and/or CMD at the end of the Dockerfile. It is generally recommended that you separate areas of concern by using one service per container. That service may fork into multiple processes (for example, Apache web server starts multiple worker processes). It’s ok to have multiple processes, but to get the most benefit out of Docker, avoid one container being responsible for multiple aspects of your overall application. You can connect multiple containers using user-defined networks and shared volumes.

The container’s main process is responsible for managing all processes that it starts. In some cases, the main process isn’t well-designed, and doesn’t handle “reaping” (stopping) child processes gracefully when the container exits. If your process falls into this category, you can use the --init option when you run the container. The --init flag inserts a tiny init-process into the container as the main process, and handles reaping of all processes when the container exits. Handling such processes this way is superior to using a full-fledged init process such as sysvinit, upstart, or systemd to handle process lifecycle within your container.

> Credit: https://docs.docker.com/config/containers/multi-service_container/

### "Distroless" Docker Images
"Distroless" images contain only your application and its runtime dependencies. They do not contain package managers, shells or any other programs you would expect to find in a standard Linux distribution.

### Hands On
In this demo, we are going to demonstrate how we use an init-system for our container with the distroless base image at the final layer.To do so, we need to use an init-system which is statically linked and also a Go binary which is also statically linked. We are going to use [Tini as a init-system](https://github.com/krallin/tini) for our container and a [static-debian distroless image](gcr.io/distroless/static-debian10) as our base image for final image.

Let's see what is inside of our Dockerfile.
```Dockerfile
# Specify base image
FROM golang:1.15.7-alpine as builder

# Specify working directory
WORKDIR /app

# Add Tini init-system
ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-static /tini
RUN chmod +x /tini

# Define environment variables for go build time
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Efficient cache usage
COPY go.mod go.sum ./
RUN go mod download

# Copy everything from host to the image
COPY . .

# Build statically compiled binary
RUN go build -o hello-world

FROM gcr.io/distroless/static-debian10

COPY --from=builder /app/hello-world ./
COPY --from=builder /tini /tini

ENTRYPOINT ["/tini", "--"]
CMD ["./hello-world"]
```
You should notice that we are using a tini-static binary instead of tini because we are using an static-debian image for our final image,so we need a statically linked version of tini, also, we disable the CGO using CGO_ENABLED environment variable to be able to build statically linked Go binary.

Then look inside of the Go program, it is very straightforward. We'll start a sleep process then we'll display the parent and child  relationship of it using [go-ps package](https://github.com/mitchellh/go-ps).

```golang
package main

import (
	"log"
	"time"

	"github.com/mitchellh/go-ps"
)

func main() {
	list, err := ps.Processes()
	if err != nil {
		panic(err)
	}
	for _, p := range list {
		log.Printf("Process %s with PID %d and PPID %d", p.Executable(), p.Pid(), p.PPid())
	}

	time.Sleep(3600 * time.Second)
}

```

Let's build our docker image and start the container.
```bash
$ docker buildx build -t tini-with-distroless:0.0.1 .
[+] Building 1.8s (18/18) FINISHED
=> [internal] load build definition from Dockerfile     0.0s                                                                    
=> transferring dockerfile: 32B                         0.0s
=> [internal] load .dockerignore                        0.0s                                                                        ...

$ docker container run tini-with-distroless:0.0.1
2021/01/31 10:50:57 Process tini with PID 1 and PPID 0
2021/01/31 10:50:57 Process hello-world with PID 8 and PPID 1
```
You should see the similar output above.
