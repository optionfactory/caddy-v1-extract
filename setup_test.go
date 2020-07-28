package extract

import (
	"github.com/caddyserver/caddy"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/caddyserver/caddy/caddyhttp/httpserver"
)

func TestParseConfig(t *testing.T) {
	controller := caddy.NewTestController("http", `
		localhost:8080
		extract regexp variable source 0
	`)
	actual, err := parseConfig(controller)
	if err != nil {
		t.Errorf("parseConfig return err: %v", err)
	}
	expected := Config{
		Regex:        regexp.MustCompile("regexp"),
		VariableName: "variable",
		Source:       "source",
		Index:        0,
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v actual %v", expected, actual)
	}
}

type testResponseRecorder struct {
	*httpserver.ResponseWriterWrapper
}

func (testResponseRecorder) CloseNotify() <-chan bool { return nil }

func TestReplacers(t *testing.T) {
	config := Config{
		Regex:        regexp.MustCompile(`^(.*?)\.`),
		VariableName: "variable",
		Source:       "{host}",
		Index:        1,
	}

	l := CaddyRegex{
		Next: httpserver.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		}),
		Config: config,
	}

	w := httptest.NewRecorder()
	recordRequest := httpserver.NewResponseRecorder(w)
	reader := strings.NewReader(``)

	r := httptest.NewRequest("GET", "http://third.level.domain", reader)
	//r.RemoteAddr = "212.50.99.193"
	rr := httpserver.NewResponseRecorder(testResponseRecorder{
		ResponseWriterWrapper: &httpserver.ResponseWriterWrapper{ResponseWriter: recordRequest},
	})

	rr.Replacer = httpserver.NewReplacer(r, rr, "-")

	l.ServeHTTP(rr, r)

	if got, want := rr.Replacer.Replace("{variable}"), "third"; got != want {
		t.Errorf("Expected custom placeholder {variable} to be set (%s), but it wasn't; got: %s", want, got)
	}

}
