package builtin

import (
	"net/http"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/trusch/bobbyd/middleware"
	"github.com/trusch/bobbyd/middleware/registry"
	"github.com/vulcand/oxy/trace"
)

func traceConstructor(next http.Handler, options interface{}) (middleware.Middleware, error) {
	opts := &traceOpts{}
	err := mapstructure.Decode(options, opts)
	if err != nil {
		return nil, err
	}
	if opts.Output == "" {
		opts.Output = "/dev/stdout"
	}
	w, err := os.Create(opts.Output)
	if err != nil {
		return nil, err
	}
	return trace.New(next, w)
}

type traceOpts struct {
	Output string
}

func init() {
	registry.Register("trace", traceConstructor)
}
