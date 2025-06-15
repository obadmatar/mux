package mux

import (
	"net/http"
	"sync"
	"time"
)

// App denotes the Mux application.
// It holds global server state, configuration, and the HTTP server instance.
type App struct {
	// config holds the application-wide configuration.
	config Config

	// pool is used to manage reusable Context instances for performance.
	pool sync.Pool

	// mutex protects mutable App state during configuration or shutdown.
	mutex sync.Mutex

	// server is the underlying HTTP server.
	server *http.Server

	// mux is the HTTP request multiplexer for routing
	mux *http.ServeMux

	// middleware holds the global middleware stack
	middleware []MiddlewareFunc
}

// Config is a struct holding the server settings.
type Config struct {
	// Max body size that the server accepts.
	// -1 will decline any body size
	//
	// Default: 4 * 1024 * 1024
	BodyLimit int `json:"body_limit"`

	// ReadTimeout is the maximum duration for reading the entire request, including the body.
	// A zero value means no timeout is set by the server.
	//
	// Default: 15s
	ReadTimeout time.Duration `json:"read_timeout"`

	// WriteTimeout is the maximum duration before timing out writes of the response.
	// A zero value means no timeout is set by the server.
	//
	// Default: 15s
	WriteTimeout time.Duration `json:"write_timeout"`

	// IdleTimeout is the maximum time to wait for the next request on a keep-alive connection.
	// A zero value means idle connections are never closed automatically.
	//
	// Default: 60s
	IdleTimeout time.Duration `json:"idle_timeout"`

	// ErrorHandler is executed when an error is returned from fiber.Handler.
	//
	// Default: DefaultErrorHandler
	ErrorHandler ErrorHandler `json:"-"`
}

// New creates a new Mux application with the given configuration.
// Zero values in config will be replaced with sensible defaults.
func New(config Config) *App {
	// Apply default body size if not explicitly set.
	if config.BodyLimit == 0 {
		config.BodyLimit = 4 * 1024 * 1024
	}
	// Apply default timeouts if unset.
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 15 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 15 * time.Second
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = 60 * time.Second
	}
	// Assign default error handler if none provided.
	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultErrorHandler
	}

	app := &App{
		config: config,

		// Initialize the context pool to reduce allocations on each request.
		pool: sync.Pool{
			New: func() interface{} {
				// Replace with real context init if it needs more setup later.
				return new(Context)
			},
		},

		// Initialize routing components
		mux:        http.NewServeMux(),
		middleware: make([]MiddlewareFunc, 0),
	}

	// Create HTTP server with the app as the handler
	app.server = &http.Server{
		Handler:      app, // Set the app as the handler immediately
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return app
}
