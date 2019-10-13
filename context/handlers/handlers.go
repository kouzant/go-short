package handlers

import (
	"net/http"
	"net/url"
	"fmt"
	"strings"
	_ "errors"
	
	"github.com/kouzant/go-short/storage"
)

/**
 * HTTP handler for redirecting requests
*/
type RedirectHandler struct {
	StateStore storage.StateStore
}

func (h *RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tokens := strings.SplitAfterN(r.URL.Path, "/", 2)
	value, error := h.StateStore.Load(storage.StorageKey(tokens[1]))
	if error != nil {
		fmt.Fprintf(w, "Error: %v", error)
	} else {
		http.Redirect(w, r, value.(string), http.StatusTemporaryRedirect)
	}	
}

/**
 * HTTP handler for administrative tasks
*/

type AdminHandler struct {
	StateStore storage.StateStore
}

func (h *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	command, err := parseAdminOp(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
	}

	switch command.(type) {
	case AddCommand:
		add, _ := command.(AddCommand)
		h.handleAddCommand(add, w)
	case DeleteCommand:
		delete := command.(DeleteCommand)
		h.handleDeleteCommand(delete, w)
	case ListCommand:
		list := command.(ListCommand)
		h.handleListCommand(list, w)
	}
}

func (h *AdminHandler) handleAddCommand(command AddCommand, w http.ResponseWriter) {
	err := h.StateStore.Save(storage.NewStorageItem(command.key, command.url))
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Added <%s, %s> to store", command.key, command.url)	
}

func (h *AdminHandler) handleDeleteCommand(command DeleteCommand, w http.ResponseWriter) {
	value, err := h.StateStore.Delete(storage.StorageKey(command.key))
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Deleted key %s", value)	
}

func (h *AdminHandler) handleListCommand(command ListCommand, w http.ResponseWriter) {
	fmt.Println("List command")
}

type AdminCommand interface{}

type AddCommand struct {
	key string
	url string
}

type DeleteCommand struct {
	key string
}

type ListCommand struct {
}



func parseAdminOp(r *http.Request) (AdminCommand, error) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil, err
	}
	switch r.Method {
	case "POST":
		// Add
		key := values.Get("key")
		if key == "" {
			return nil, fmt.Errorf("Add command is missing key parameter")
		}
		url := values.Get("url")
		if url == "" {
			return nil, fmt.Errorf("Add command is missing url parameter")
		}
		return AddCommand{key, url}, nil
	case "DELETE":
		// Delete
		key := values.Get("key")
		if key == "" {
			return nil, fmt.Errorf("Delete command is missing key parameter")
		}
		return DeleteCommand{key}, nil
	case "GET":
		// List all
		return ListCommand{}, nil
	default:
		return nil, fmt.Errorf("Unknown method")
	}
}
