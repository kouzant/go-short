package main

import (
	"flag"
	"os"
	"fmt"
	"net/http"
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
	opArg := clientMode.String("operation", "add", "Operation (add | delete | list)")
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
		redirectHandler := &handlers.RedirectHandler{StateStore: stateStore}
		adminHandler := &handlers.AdminHandler{StateStore: stateStore}
		mux.Handle("/", redirectHandler)
		mux.Handle("/_admin", adminHandler)
		
		listeningOn := fmt.Sprintf("%s:%d", conf.GetString(context.WebListenKey),
			conf.GetInt(context.WebPortKey))
		log.Info("Start listening on ", listeningOn)
		log.Fatal(http.ListenAndServe(listeningOn, mux))
	} else if clientMode.Parsed() {
		switch *opArg {
		case "add":
			if *keyArg == "" || *valueArg == "" {
				clientMode.PrintDefaults()
				os.Exit(1)
			}
			err := stateStore.Save(storage.NewStorageItem(*keyArg, *valueArg))
			if err != nil {
				fmt.Printf("> ERROR: Save operation could not complete, reason: %s\n", err)
				os.Exit(3)
			}
			fmt.Println("> Added <", *keyArg, ", ", *valueArg, "> to go-short!")
		case "delete":
			if *keyArg == "" {
				clientMode.PrintDefaults()
				os.Exit(2)
			}
			value, err := stateStore.Delete(storage.StorageKey(*keyArg))
			if err != nil {
				fmt.Printf("> ERROR: Could not delete key %s reason: %s\n", *keyArg, err)
				os.Exit(3)
			}
			fmt.Println("> Deleted ", value, " from go-short!")
		case "list":
			storedItems, err := stateStore.LoadAll()
			if err != nil {
				fmt.Printf("> ERROR: Could not list all URLs, reason %s\n", err)
				os.Exit(3)
			}
			fmt.Printf("> Number of shortened URLs: %d\n", len(storedItems))
			for _, item := range storedItems {
				fmt.Printf("> Short: %s\t URL: %s\n", item.Key, item.Value)
			}
		default:
			clientMode.PrintDefaults()
			os.Exit(1)
		}
	}
}
