## go-short

`go-short` is a simple URL shortener for personal usage. It starts an HTTP server and redirects shortened URLs to their
associated normal URL.


### Table of contents
1. [Installation](#installation)
2. [Configuration](#configuration)
3. [Usage](#usage)
4. [Development](#development)

### Installation
To install `go-short` download the binary from [here](https://github.com/kouzant/go-short/releases). You can start the server with `./go-short server`. In the _resources/_ folder there is a sample systemd unit file if you want to start it through systemd.

In case you want the server to listen on a privileged port you should set the special capabilities for the binary file.

    setcap 'cap_net_bind_service=+ep' /usr/local/go-short/go-short

`authbind` won't work as the binary is statically linked.

#### Docker
You can also run `go-short` server in a Docker container instead of running it directly in your host. Follow the instructions
on [Docker Hub](https://hub.docker.com/r/kouzan/go-short).

### Configuration
Both the server and the client will look for a configuration file in `$HOME/.go-short/go-short.yml` Configuration is the following

    go-short:
      # application logging level
      log-level: info
      state-store:
       # path where go-short will save its state
       path: /home/antonis/.go-short/state_store
       # How often will we perfom GC on the state store
       gc-interval: 2h
      webserver:
       # IP the HTTP server will listen to
       listen: 127.0.0.1
       # port the server will listen to
       port: 80

You will also need to change _/etc/hosts_ so that **go** (or anything else) domain name will resolve to localhost.
It should look like the following:

    127.0.0.1	localhost go

### Usage
There two main commands, `server` which will start the server and `client` which performs operations on the server such as
`add`, `delete` and `list`. `server` mode does not take any other arguments. `client` has the following sub-commands:

    ./go-short client
      -file string
    	    Path to CSV file key,URL
      -key string
    	    Shortened URL key
      -op string
    	    Operation (add | delete | list | add-batch) (default "add")
      -url string
    	    URL
          
* To add a new short URL type `./go-short client -key gs -url https://github.com/kouzant/go-short`
* To list all shortened URLs type `./go-short client -op list` or use the web UI shown below
* To delete a URL type `./go-short client -op delete -key gs`
* To add batch entries from a CSV file type `./go-short client -op add-batch -file FILE_PATH`

After you've added a short URL, go to your browser and type `go/gs`. It will redirect you to [https://github.com/kouzant/go-short](https://github.com/kouzant/go-short)

You can also list the shortened URLs in a nicer(?) way by visiting `go/_admin`

### Development
`go-short` is written in Go 1.13 and is using [Badger](https://github.com/dgraph-io/badger) as a persistent state store. To run all the tests execute `go test ./...` or if you want the tests of a specific package
e.g. `go test github.com/kouzant/go-short/storage`

To buld it run `go build`

To build the Docker image run `docker build -t kouzan/go-short:VERSION -f resources/Dockerfile .`
