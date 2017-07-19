package main

import (
	"flag"
	"net/http"
)

func main() {
	var (
		err        error
		controller *Controller
		log        *Log
		router     *Router
	)

	// Read flags
	flag.Parse()
	// Create logger
	log = NewLogger()
	// Print version if flag passed
	showVersion(log)

	if controller, err = NewController(log); err != nil {
		log.Fatal(err)
	}

	// Create router
	router = NewRouter()

	// Login
	router.Handle("PUT", "/login", controller.Handle(controller.Login))
	// Get users (mailboxes)
	router.Handle("GET", "/users", controller.Handle(controller.Users))
	router.Handle("GET", "/user/:uid", controller.Handle(controller.User))

	// Spamm
	router.Handle("GET", "/spam", controller.Handle(controller.Spam))

	// Transport
	router.Handle("GET", "/transports", controller.Handle(controller.Transports))
	router.Handle("GET", "/transport/:tid", controller.Handle(controller.Transports))

	// Handle NotFound
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Notice("%s: 404 (Not Found)", r.Context().Value("Id"))
		http.NotFound(w, r)
	})

	http.ListenAndServe(SERVERADDRESS, Middlewares(
		router,
		JWT(log),
		Assets(ASSETSPATH),
		// Keep this middleware last
		RequestId(log),
	))
}
