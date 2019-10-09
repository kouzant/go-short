package main

import (
	"flag"
	"os"
	"fmt"
	"net/http"
	"strings"
	
	"github.com/kouzant/go-short/context"
	"github.com/kouzant/go-short/logger"
	"github.com/kouzant/go-short/storage"	
	log "github.com/sirupsen/logrus"	
)


func main() {
	serverMode := flag.NewFlagSet("server", flag.ExitOnError)
	clientMode := flag.NewFlagSet("client", flag.ExitOnError)
	
	// Client mode arguments
	opArg := clientMode.String("operation", "add", "Operation (add | delete)")
	keyArg := clientMode.String("key", "", "Shortened URL key")
	valueArg := clientMode.String("url", "", "URL")

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

	stateStore := &storage.MemoryStateStore{Config: conf}
	stateStore.Init()
	defer stateStore.Close()

	item := &storage.StorageItem{
		storage.StorageKey("lala"),
		storage.StorageValue("https://logicalclocks.com")}
	stateStore.Save(item)
	
	if serverMode.Parsed() {
		fmt.Println("Server parsed")
		mux := http.NewServeMux()
		rh := &RedirectHandler{stateStore: stateStore}
		mux.Handle("/", rh)
		listeningOn := fmt.Sprintf("%s:%d", conf.GetString(context.WebListenKey),
			conf.GetInt(context.WebPortKey))
		http.ListenAndServe(listeningOn, mux)
	} else if clientMode.Parsed() {
		fmt.Println("Client parsed")		
		switch *opArg {
		case "add":
			fmt.Println("op add")
			if *keyArg == "" || *valueArg == "" {
				clientMode.PrintDefaults()
				os.Exit(1)
			}
			fmt.Printf("op: add key: %s url: %s\n", *keyArg, *valueArg)
		case "delete":
			fmt.Println("op delete")
			if *keyArg == "" {
				clientMode.PrintDefaults()
				os.Exit(2)
			}
			fmt.Printf("op delete key: %s\n", *keyArg)
		default:
			clientMode.PrintDefaults()
			os.Exit(1)
		}
	}
}

type RedirectHandler struct {
	stateStore storage.StateStore
}

func (h *RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tokens := strings.SplitAfterN(r.URL.Path, "/", 2)
	value, error := h.stateStore.Load(storage.StorageKey(tokens[1]))
	if error != nil {
		fmt.Fprintf(w, "Error: %v", error)
	} else {
		http.Redirect(w, r, value.(string), http.StatusMovedPermanently)
	}
}
