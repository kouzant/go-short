package handlers

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/kouzant/go-short/storage"
)

func TestCommandParsing(t *testing.T) {
	baseUrl := "http://go/_admin?"
	shortenUrl := "https://github.com/kouzant/go-short"
	var tests = []struct {
		params string
		body   string
		method string
		want   AdminCommand
	}{
		{"key=gs&url=" + shortenUrl, "", "POST", AddCommand{"gs", shortenUrl}},
		{"url=" + shortenUrl, "", "POST", nil},
		{"key=gs", "", "POST", nil},
		{"", "", "POST", nil},

		{"key=gs", "", "DELETE", DeleteCommand{"gs"}},
		{"", "", "DELETE", nil},

		{"", "", "GET", ListCommand{}},

		{"", "key0,val0\nkey1,val1", "PUT", AddBatchCommand{[]*storage.Pair{&storage.Pair{Left: "key0", Right: "val0"}, &storage.Pair{Left: "key1", Right: "val1"}}}},
	}

	for _, test := range tests {
		reqUrl := baseUrl + test.params
		r, err := http.NewRequest(test.method, reqUrl, string2Reader(test.body))
		if err != nil {
			t.Errorf("Error creating new HTTP request %s", err)
		}

		command, err := parseAdminOp(r)
		if err != nil && command != test.want {
			t.Errorf("parsingAdminOp(%v) return error %v", r, err)
		}

		switch command.(type) {
		case AddBatchCommand:
			if !compareAddBatchCommand(command.(AddBatchCommand), test.want.(AddBatchCommand)) {
				t.Errorf("Expected parsed AddBatchCommand %v to equal %v", command, test.want)
			}
		default:
			if command != test.want {
				t.Errorf("parsingAdminOp(%v) Expected %v gotten %v", r, test.want, command)
			}
		}
	}
}

func compareAddBatchCommand(command, want AddBatchCommand) bool {
	for _, wantPair := range want.pairs {
		pairFound := false
		for _, commandPair := range command.pairs {
			if wantPair.Left == commandPair.Left && wantPair.Right == commandPair.Right {
				pairFound = true
				break
			}
		}
		if !pairFound {
			return false
		}
	}
	return true
}
func string2Reader(str string) io.Reader {
	if str == "" {
		return nil
	}
	return strings.NewReader(str)
}
