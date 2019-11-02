package handlers

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/kouzant/go-short/context"
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

var list_all_template = template.Must(template.New("list").Parse(list_all_html))

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
		h.handleListCommand(list, w, r.UserAgent())
	case AddBatchCommand:
		addBatch := command.(AddBatchCommand)
		h.handleAddBatchCommand(addBatch, w)
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

func (h *AdminHandler) handleAddBatchCommand(command AddBatchCommand, w http.ResponseWriter) {
	if len(command.pairs) == 0 {
		http.Error(w, "No parameters passed", http.StatusBadRequest)
		return
	}
	items := make([]*storage.StorageItem, 0, len(command.pairs))
	for _, p := range command.pairs {
		tokens := strings.Split(p, ",")
		if len(tokens) != 2 {
			continue
		}
		items = append(items, storage.NewStorageItem(tokens[0], tokens[1]))
	}
	if len(items) == 0 {
		http.Error(w, "No parameters passed", http.StatusBadRequest)
		return
	}
	err := h.StateStore.SaveAll(items)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "Added pairs to store")
}

func (h *AdminHandler) handleDeleteCommand(command DeleteCommand, w http.ResponseWriter) {
	value, err := h.StateStore.Delete(storage.StorageKey(command.key))
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Deleted key %s", value)
}

func (h *AdminHandler) handleListCommand(command ListCommand, w http.ResponseWriter,
	userAgent string) {
	storedItems, err := h.StateStore.LoadAll()
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}
	// If is's the CLI return simple string
	// otherwise template an HTML page
	if userAgent == context.CLI_USER_AGENT {
		var buffer strings.Builder
		fmt.Fprintf(&buffer, "> Number of stored items: %d\n", len(storedItems))

		for _, item := range storedItems {
			fmt.Fprintf(&buffer, "> Short: %s\t URL: %s\n", item.Key, item.Value)
		}
		fmt.Fprint(w, buffer.String())
	} else {
		err := list_all_template.Execute(w, storedItems)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		}
	}
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

type AddBatchCommand struct {
	pairs []string
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
	case "PUT":
		// Add batch
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("Add batch request is missing body")
		}
		pairs := strings.Split(strings.TrimSpace(string(body)), "\n")
		return AddBatchCommand{pairs: pairs}, nil
	default:
		return nil, fmt.Errorf("Unknown method")
	}
}

const list_all_html = `
<html>
 <head>
   <style>
     table {
     font-family: arial, sans-serif;
     border-collapse: collapse;
     width: 80%;
     }

     td, th {
     border: 1px solid #dddddd;
     text-align: center;
     padding: 8px;
     }

     .short {
     border: 1px solid #dddddd;
     text-align: center;
     padding: 8px;
     }

     .long {
     border: 1px solid #dddddd;
     text-align: left;
     padding: 8px;
     }

     tr:nth-child(even) {
     background-color: #dddddd;
     }
   </style>
 </head>
 <body>
    <div align="center">
    <h1>go-shortened URLs: {{len .}}</h1>
    <h3>If you have no idea what's this, go check project's <a href="https://github.com/kouzant/go-short" target="_blank">GitHub page</a></h3>
    <table>
      <tr>
	<th>Shortened</th>
	<th>URL</th>
      </tr>

{{range .}}
      <tr>
	<td class="short">{{.Key}}</td>
	<td class="long"><a href="{{.Value}}">{{.Value}}</a></td>
      </tr>
{{end}}
    </table>
    </div>
  </body>
</html>
`
