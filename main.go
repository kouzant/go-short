package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kouzant/go-short/context"
	"github.com/kouzant/go-short/context/handlers"
	"github.com/kouzant/go-short/logger"
	"github.com/kouzant/go-short/storage"
	log "github.com/sirupsen/logrus"
)

func main() {
	serverMode := flag.NewFlagSet("server", flag.ExitOnError)
	clientMode := flag.NewFlagSet("client", flag.ExitOnError)

	// Client mode arguments
	opArg := clientMode.String("op", "add", "Operation (add | delete | list | add-batch)")
	keyArg := clientMode.String("key", "", "Shortened URL key")
	valueArg := clientMode.String("url", "", "URL")
	batchFileArg := clientMode.String("file", "", "Path to CSV file key,URL")

	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s [server | client] ...\n", os.Args[0])
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		serverMode.Parse(os.Args[2:])
	case "client":
		clientMode.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	conf := context.ReadConfig()
	logger.Init(conf)
	log.Info("Starting go-short")

	listeningOn := fmt.Sprintf("%s:%d", conf.GetString(context.WebListenKey),
		conf.GetInt(context.WebPortKey))
	if serverMode.Parsed() {
		stateStore := &storage.BadgerStateStore{Config: conf}
		error := stateStore.Init()
		if error != nil {
			log.Fatal("Could not initialize state store ", error)
		}
		defer stateStore.Close()

		// Trap exit signal
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-sigs
			log.Infof("Received %s\n", sig)
			if error := stateStore.Close(); error != nil {
				log.Errorf("Error closing the state store %s\n", error)
				os.Exit(2)
			}
			log.Info("Bye...")
			os.Exit(0)
		}()

		mux := http.NewServeMux()
		redirectHandler := &handlers.RedirectHandler{StateStore: stateStore}
		adminHandler := &handlers.AdminHandler{StateStore: stateStore}
		mux.Handle("/", redirectHandler)
		mux.Handle("/_admin", adminHandler)

		log.Info("Start listening on ", listeningOn)
		log.Fatal(http.ListenAndServe(listeningOn, mux))
	} else if clientMode.Parsed() {
		switch *opArg {
		case "add":
			if *keyArg == "" || *valueArg == "" {
				clientMode.PrintDefaults()
				os.Exit(1)
			}
			doAddRequest(listeningOn, *keyArg, *valueArg)
		case "delete":
			if *keyArg == "" {
				clientMode.PrintDefaults()
				os.Exit(1)
			}
			doDeleteRequest(listeningOn, *keyArg)
		case "list":
			doListRequest(listeningOn)
		case "add-batch":
			if *batchFileArg == "" {
				fmt.Printf("> ERROR: Missing -file argument")
				clientMode.PrintDefaults()
				os.Exit(1)
			}
			doBatchAddRequest(listeningOn, *batchFileArg)
		default:
			clientMode.PrintDefaults()
			os.Exit(1)
		}
	}
}

func doAddRequest(url, key, value string) {
	reqUrl := fmt.Sprintf("http://%s/_admin?key=%s&url=%s", url, key, value)
	statusCode, body := doRequest("POST", reqUrl, nil)

	if statusCode == http.StatusOK {
		fmt.Println(string(body))
	} else {
		fmt.Printf("> ERROR: %s\n", body)
		os.Exit(3)
	}
}

func doDeleteRequest(url, key string) {
	reqUrl := fmt.Sprintf("http://%s/_admin?key=%s", url, key)
	statusCode, body := doRequest("DELETE", reqUrl, nil)

	if statusCode == http.StatusOK {
		fmt.Println(string(body))
	} else {
		fmt.Printf("> ERROR: %s\n", body)
		os.Exit(3)
	}
}

func doListRequest(url string) {
	reqUrl := fmt.Sprintf("http://%s/_admin", url)
	statusCode, body := doRequest("GET", reqUrl, nil)

	if statusCode == http.StatusOK {
		fmt.Println(string(body))
	} else {
		fmt.Printf("> ERROR: %s\n", body)
		os.Exit(3)
	}
}

func doBatchAddRequest(url, path string) {
	fd, err := os.Open(path)
	if err != nil {
		fmt.Printf("> ERROR: Could not open file %s\n", err)
	}
	defer fd.Close()
	scanner := bufio.NewScanner(fd)
	var b bytes.Buffer
	for scanner.Scan() {
		text := fmt.Sprintf("%s\n", scanner.Text())
		b.WriteString(text)
	}

	reqUrl := fmt.Sprintf("http://%s/_admin", url)
	var r io.Reader
	r = &b
	statusCode, body := doRequest("PUT", reqUrl, r)

	if statusCode == http.StatusOK {
		fmt.Println(string(body))
	} else {
		fmt.Printf("> ERROR: %s\n", body)
		os.Exit(3)
	}
}

func doRequest(method, url string, reqBody io.Reader) (int, []byte) {
	client := http.Client{}
	req, err := http.NewRequest(method, url, reqBody)
	handleClientError(method, err)
	req.Header.Add("User-Agent", context.CLI_USER_AGENT)
	resp, err := client.Do(req)
	handleClientError("add", err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	handleClientError(method, err)

	return resp.StatusCode, body
}

func handleClientError(op string, err error) {
	if err != nil {
		fmt.Printf("> ERROR: Could not %s item, reason %s\n", op, err)
		os.Exit(3)
	}
}
