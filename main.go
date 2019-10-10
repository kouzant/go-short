package main

import (
	"flag"
	"os"
	"fmt"
	"net/http"
	"strings"
	"os/signal"
	"syscall"
	
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

	stateStore := &storage.BadgerStateStore{Config: conf}
	error := stateStore.Init()
	if error != nil {
		log.Fatal("Could not initialize state store ", error)
	}
	defer stateStore.Close()

	if serverMode.Parsed() {
		// Trap exit signal
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func(){
			sig := <- sigs
			log.Infof("Received %s\n", sig)
			if error := stateStore.Close(); error != nil {
				log.Errorf("Error closing the state store %s\n", error)
				os.Exit(2)
			}
			log.Info("Bye...")
			os.Exit(0)
		}()
		
		mux := http.NewServeMux()
		rh := &RedirectHandler{stateStore: stateStore}
		mux.Handle("/", rh)
		listeningOn := fmt.Sprintf("%s:%d", conf.GetString(context.WebListenKey),
			conf.GetInt(context.WebPortKey))
		log.Info("Start listening on ", listeningOn)
		log.Fatal(http.ListenAndServe(listeningOn, mux))
	} else if clientMode.Parsed() {
		fmt.Println("Client parsed")		
		switch *opArg {
		case "add":
			fmt.Println("op add")
			if *keyArg == "" || *valueArg == "" {
				clientMode.PrintDefaults()
				os.Exit(1)
			}
			err := stateStore.Save(storage.NewStorageItem(*keyArg, *valueArg))
			if err != nil {
				fmt.Println("> ERROR: Save operation could not complete, reason: ", err)
			}
			fmt.Println("> Added <", *keyArg, ", ", *valueArg, "> to go-short!")
		case "delete":
			fmt.Println("op delete")
			if *keyArg == "" {
				clientMode.PrintDefaults()
				os.Exit(2)
			}
			value, err := stateStore.Delete(storage.StorageKey(*keyArg))
			if err != nil {
				fmt.Println("> ERROR: Could not delete key ", *keyArg, ", reason: ", err)
			}
			fmt.Println("> Deleted <", *keyArg, ", ", value, "> from go-short!")
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
		http.Redirect(w, r, value.(string), http.StatusTemporaryRedirect)
	}
}
