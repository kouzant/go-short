package handlers

import (
	"net/http"
	"testing"
)

func TestCommandParsing(t *testing.T) {
	baseUrl := "http://go/_admin?"
	shortenUrl := "https://github.com/kouzant/go-short"
	var tests = []struct {
		params string
		method string
		want   AdminCommand
	}{
		{"key=gs&url=" + shortenUrl, "POST", AddCommand{"gs", shortenUrl}},
		{"url=" + shortenUrl, "POST", nil},
		{"key=gs", "POST", nil},
		{"", "POST", nil},

		{"key=gs", "DELETE", DeleteCommand{"gs"}},
		{"", "DELETE", nil},

		{"", "GET", ListCommand{}},
	}

	for _, test := range tests {
		reqUrl := baseUrl + test.params
		r, err := http.NewRequest(test.method, reqUrl, nil)
		if err != nil {
			t.Errorf("Error creating new HTTP request %s", err)
		}

		command, err := parseAdminOp(r)
		if err != nil && command != test.want {
			t.Errorf("parsingAdminOp(%v) return error %v", r, err)
		}
		if command != test.want {
			t.Errorf("parsingAdminOp(%v) Expected %v gotten %v", r, test.want, command)
		}
	}
}
