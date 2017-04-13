package registry

import (
	"fmt"
	"net/http"

	"github.com/trusch/bobbyd/middleware"
)

var (
	constructors = make(map[string]middleware.Constructor)
)

// Register registers a new middleware type
func Register(id string, constructor middleware.Constructor) {
	constructors[id] = constructor
}

// Create constructs a new middleware instance
func Create(id string, next http.Handler, options interface{}) (middleware.Middleware, error) {
	constructor, ok := constructors[id]
	if !ok {
		return nil, fmt.Errorf("no middleware type '%v'", id)
	}
	return constructor(next, options)
}
