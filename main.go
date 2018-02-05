package main

import (
	"flag"
	"net/http"
)

func main() {
	var (
		env    *Bus
		router *Router
	)

	// Read flags
	flag.Parse()
	// Create logger
	env = &Bus{
		LogIface: NewLogger(),
	}
	// Print version if flag passed
	showVersion(env)

	if err := env.openDB(nil); err != nil {
		env.Fatal(err)
	}

	// Create router
	router = NewRouter()

	// Login
	router.Handle("POST", "/login", NewHandler(
		Login,
		env,
	))

	// Get mail aliases
	router.Handle("GET", "/aliases/groups", NewHandler(
		Protect(aliasGroupWrap(Aliases)),
		env,
	))
	router.Handle("GET", "/aliases/search", NewHandler(
		Protect(MailSearch),
		env,
	))
	router.Handle("GET", "/aliases", NewHandler(
		Protect(Aliases),
		env,
	))
	router.Handle("GET", "/alias/:aid", NewHandler(
		Protect(Alias),
		env,
	))
	router.Handle("POST", "/alias", NewHandler(
		Protect(SetAlias),
		env,
	))

	// Get users (mailboxes)
	router.Handle("GET", "/users", NewHandler(
		Protect(Users),
		env,
	))
	router.Handle("GET", "/user/:uid", NewHandler(
		Protect(User),
		env,
	))

	// Spamm
	router.Handle("GET", "/spam", NewHandler(
		Protect(Spam),
		env,
	))

	// Transport
	router.Handle("GET", "/transports", NewHandler(
		Protect(Transports),
		env,
	))
	router.Handle("GET", "/transport/:tid", NewHandler(
		Protect(Transport),
		env,
	))

	// Handle NotFound
	if ASSETSPATH != "" {
		router.NotFound = http.FileServer(http.Dir(ASSETSPATH))

		env.Notice("Using file server with public=%s for unknown routes", ASSETSPATH)
	}

	http.ListenAndServe(SERVERADDRESS, Middlewares(
		router,
		JWT(env),
		verbose(env),
		RequestId(),
	))
}
