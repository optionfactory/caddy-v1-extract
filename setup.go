package extract

import (
	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
	"net/http"
	"regexp"
	"strconv"
)

// GeoIP represents a middleware instance
type CaddyRegex struct {
	Next      httpserver.Handler
	Config    *Config
}

type Config struct {
	Regex *regexp.Regexp
	VariableName string
	Source string
	Index int
}

func parseConfig(c *caddy.Controller) (*Config, error) {
	var config = Config{}
	for c.Next() {
		value := c.Val()
		switch value {
		case "extract":
			if !c.NextArg() {
				continue
			}
			config.Regex = regexp.MustCompile(c.Val())
			if !c.NextArg() {
				continue
			}
			config.VariableName = c.Val()
			if !c.NextArg() {
				continue
			}
			config.Source = c.Val()
			if !c.NextArg() {
				continue
			}
			var err error
			config.Index, err = strconv.Atoi(c.Val())
			if err != nil {
				return nil, err
			}
		}
	}
	return &config, nil
}


// Init initializes the plugin
func init() {
	caddy.RegisterPlugin("extract", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	config, err := parseConfig(c)
	if err != nil {
		return err
	}


	// Create new middleware
	newMiddleWare := func(next httpserver.Handler) httpserver.Handler {
		return &CaddyRegex{
			Next:      next,
			Config:    config,
		}
	}
	// Add middleware
	cfg := httpserver.GetConfig(c)
	cfg.AddMiddleware(newMiddleWare)

	return nil
}

func (cr CaddyRegex) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	cr.extractRegex(w, r)
	return cr.Next.ServeHTTP(w, r)
}

func (cr CaddyRegex) extractRegex(w http.ResponseWriter, r *http.Request) {
	replacer := newReplacer(r)
	value := replacer.Replace(cr.Config.Source)
	matches := cr.Config.Regex.FindStringSubmatch(value)
	if matches == nil {
		replacer.Set(cr.Config.VariableName, "")
	} else {
		replacer.Set(cr.Config.VariableName, matches[cr.Config.Index])
	}

	if rr, ok := w.(*httpserver.ResponseRecorder); ok {
		rr.Replacer = replacer
	}
}

func newReplacer(r *http.Request) httpserver.Replacer {
	return httpserver.NewReplacer(r, nil, "")
}
