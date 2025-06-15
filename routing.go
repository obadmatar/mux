package mux

import (
	"net/http"
)

// Get registers a GET route with the given path and handler.
func (app *App) Get(path string, handler Handler, middleware ...MiddlewareFunc) {
	app.addRoute("GET", path, handler, middleware...)
}

// Post registers a POST route with the given path and handler.
func (app *App) Post(path string, handler Handler, middleware ...MiddlewareFunc) {
	app.addRoute("POST", path, handler, middleware...)
}

// Put registers a PUT route with the given path and handler.
func (app *App) Put(path string, handler Handler, middleware ...MiddlewareFunc) {
	app.addRoute("PUT", path, handler, middleware...)
}

// Delete registers a DELETE route with the given path and handler.
func (app *App) Delete(path string, handler Handler, middleware ...MiddlewareFunc) {
	app.addRoute("DELETE", path, handler, middleware...)
}

// Patch registers a PATCH route with the given path and handler.
func (app *App) Patch(path string, handler Handler, middleware ...MiddlewareFunc) {
	app.addRoute("PATCH", path, handler, middleware...)
}

// Head registers a HEAD route with the given path and handler.
func (app *App) Head(path string, handler Handler, middleware ...MiddlewareFunc) {
	app.addRoute("HEAD", path, handler, middleware...)
}

// Options registers an OPTIONS route with the given path and handler.
func (app *App) Options(path string, handler Handler, middleware ...MiddlewareFunc) {
	app.addRoute("OPTIONS", path, handler, middleware...)
}

// Use adds middleware to the application.
// Middleware will be applied to all routes registered after this call.
func (app *App) Use(middleware ...MiddlewareFunc) {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	if app.middleware == nil {
		app.middleware = make([]MiddlewareFunc, 0)
	}
	app.middleware = append(app.middleware, middleware...)
}

// Group creates a new route group with optional middleware.
// This allows for organizing routes and applying middleware to specific groups.
func (app *App) Group(prefix string, middleware ...MiddlewareFunc) *Group {
	return &Group{
		app:        app,
		prefix:     prefix,
		middleware: middleware,
	}
}

// addRoute is an internal method that registers a route with the ServeMux.
func (app *App) addRoute(method, path string, handler Handler, middleware ...MiddlewareFunc) {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	// Create the route pattern for ServeMux (method + path)
	pattern := method + " " + path

	// Wrap the handler to work with http.ServeMux
	app.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// Get a context from the pool
		ctx := app.acquireContext(r, w)
		defer app.releaseContext(ctx)

		// Apply route-specific middleware first, then global middleware
		finalHandler := handler

		// Apply route-specific middleware (in reverse order)
		for i := len(middleware) - 1; i >= 0; i-- {
			finalHandler = middleware[i](finalHandler)
		}

		// Apply global middleware
		finalHandler = app.applyMiddleware(finalHandler)

		// Execute the handler
		if err := finalHandler.Handle(ctx); err != nil {
			// Use the configured error handler
			app.config.ErrorHandler(ctx, err)
		}
	})
}

// applyMiddleware applies all registered middleware to a handler.
func (app *App) applyMiddleware(handler Handler) Handler {
	// Apply middleware in reverse order (last registered, first executed)
	for i := len(app.middleware) - 1; i >= 0; i-- {
		handler = app.middleware[i](handler)
	}
	return handler
}

// acquireContext gets a Context from the pool and initializes it.
func (app *App) acquireContext(req *http.Request, res http.ResponseWriter) *Context {
	ctx := app.pool.Get().(*Context)
	ctx.app = app
	ctx.req = req
	ctx.res = res
	return ctx
}

// releaseContext returns a Context to the pool after cleaning it.
func (app *App) releaseContext(ctx *Context) {
	// Clear references to prevent memory leaks
	ctx.app = nil
	ctx.req = nil
	ctx.res = nil
	app.pool.Put(ctx)
}

// ServeHTTP implements http.Handler interface, making App compatible with http.Server.
func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if app.mux != nil {
		app.mux.ServeHTTP(w, r)
	} else {
		// If no routes registered, return 404
		http.NotFound(w, r)
	}
}

// Listen starts the HTTP server on the specified address.
func (app *App) Listen(addr string) error {
	app.server.Addr = addr
	return app.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (app *App) Shutdown() error {
	return app.server.Shutdown(nil)
}

// Group represents a route group with shared prefix and middleware.
type Group struct {
	app        *App
	prefix     string
	middleware []MiddlewareFunc
}

// Get registers a GET route in this group.
func (g *Group) Get(path string, handler Handler, middleware ...MiddlewareFunc) {
	g.addRoute("GET", path, handler, middleware...)
}

// Post registers a POST route in this group.
func (g *Group) Post(path string, handler Handler, middleware ...MiddlewareFunc) {
	g.addRoute("POST", path, handler, middleware...)
}

// Put registers a PUT route in this group.
func (g *Group) Put(path string, handler Handler, middleware ...MiddlewareFunc) {
	g.addRoute("PUT", path, handler, middleware...)
}

// Delete registers a DELETE route in this group.
func (g *Group) Delete(path string, handler Handler, middleware ...MiddlewareFunc) {
	g.addRoute("DELETE", path, handler, middleware...)
}

// Patch registers a PATCH route in this group.
func (g *Group) Patch(path string, handler Handler, middleware ...MiddlewareFunc) {
	g.addRoute("PATCH", path, handler, middleware...)
}

// Head registers a HEAD route in this group.
func (g *Group) Head(path string, handler Handler, middleware ...MiddlewareFunc) {
	g.addRoute("HEAD", path, handler, middleware...)
}

// Options registers an OPTIONS route in this group.
func (g *Group) Options(path string, handler Handler, middleware ...MiddlewareFunc) {
	g.addRoute("OPTIONS", path, handler, middleware...)
}

// Use adds middleware to this group.
func (g *Group) Use(middleware ...MiddlewareFunc) {
	g.middleware = append(g.middleware, middleware...)
}

// Group creates a sub-group with additional prefix and middleware.
func (g *Group) Group(prefix string, middleware ...MiddlewareFunc) *Group {
	return &Group{
		app:        g.app,
		prefix:     g.prefix + prefix,
		middleware: append(g.middleware, middleware...),
	}
}

// addRoute adds a route to the group with the group's prefix and middleware.
func (g *Group) addRoute(method, path string, handler Handler, middleware ...MiddlewareFunc) {
	fullPath := g.prefix + path

	// Combine group middleware with route-specific middleware
	allMiddleware := make([]MiddlewareFunc, 0, len(g.middleware)+len(middleware))
	allMiddleware = append(allMiddleware, g.middleware...)
	allMiddleware = append(allMiddleware, middleware...)

	g.app.addRoute(method, fullPath, handler, allMiddleware...)
}
