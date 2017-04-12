package middleware

import "net/http"

// Middleware is the middleware type (simple http.Handler)
type Middleware http.Handler

// Constructor is the signature of a middleware constructor
type Constructor func(next http.Handler, options interface{}) (Middleware, error)
