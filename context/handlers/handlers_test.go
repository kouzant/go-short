package handlers

import (
	"testing"
	"net/url"
)

func TestCommandParsing(t *testing.T) {
	baseUrl := "http://go/_admin?"
	shortenUrl := "https://github.com/kouzant/go-short"
	var tests = []struct{
		params string
		want AdminCommand
	}{
		{"op=add&key=gs&url=" + shortenUrl, AddCommand{"gs", shortenUrl}},
		{"key=gs&url=" + shortenUrl, nil},
		{"op=add&url=" + shortenUrl, nil},
		{"op=add&key=gs", nil},

		{"op=delete&key=gs", DeleteCommand{"gs"}},
		{"key=gs", nil},
		{"op=add", nil},

		{"op=list", ListCommand{}},
		{"", nil},
	}

	for _, test := range tests {
		reqUrl, err := url.Parse(baseUrl + test.params)
		if err != nil {
			t.Errorf("Error parsing url")
		}
		command, err := parseAdminOp(reqUrl)
		if err != nil && command != test.want {
			t.Errorf("parsingAdminOp(%v) return error %v", reqUrl, err)
		}
		if command != test.want {
			t.Errorf("Expected %v gotten %v", test.want, command)
		}
	}
}
