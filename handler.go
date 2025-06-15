package mux

import (
	"log"
	"net/http"
)

// Handler defines an interface for handling HTTP requests.
// Custom handlers must implement this interface.
// Handle receives a Context and returns an error if the processing fails.
type Handler interface {
	Handle(ctx *Context) error
}

// HandlerFunc is an adapter to use ordinary functions as HTTP handlers.
type HandlerFunc func(ctx *Context) error

// Handle implements the Handler interface for HandlerFunc.
// It simply calls the underlying function.
func (f HandlerFunc) Handle(ctx *Context) error {
	return f(ctx)
}

// MiddlewareFunc defines a function to process middleware.
// Middleware wraps a Handler to provide additional processing.
type MiddlewareFunc func(Handler) Handler

// ErrorHandler defines a function that will process all errors
// returned from any handler in the stack
type ErrorHandler = func(*Context, error) error

// DefaultErrorHandler is the fallback error handler used if none is provided in Config.
// It sends a 500 Internal Server Error with a generic message to the client,
// and logs the detailed error for server-side visibility.
var DefaultErrorHandler ErrorHandler = func(c *Context, err error) error {
	// Defensive: nil Context or nil response writer should never happen, but avoid panic if so.
	if c == nil || c.res == nil {
		log.Printf("error: %v (context nil)", err)
		return err
	}

	// Log the error. In production, this might go to a structured logger with request metadata.
	log.Printf("internal server error: %v", err)

	// Write generic 500 response. Avoid exposing internal error messages to the client.
	http.Error(
		c.res,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)

	return err
}

// Context represents the Context which hold the HTTP request and response.
// It has methods for the request query string, parameters, body, HTTP headers and so on.
type Context struct {
	// app is a reference to the parent App instance.
	app *App

	// req is the underlying HTTP request.
	req *http.Request

	// res is the HTTP response writer.
	res http.ResponseWriter
}
