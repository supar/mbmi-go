package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

var (
	// Program name
	programName,
	// Program version
	programVersion,
	// Listen
	SERVERADDRESS,
	// Assets
	ASSETSPATH,
	// JWT secret
	SECRETPHRASE,
	// Build date and time
	buildDate,
	// Database user
	DBUSER,
	// Database user password
	DBPASS,
	// Database name
	DBNAME,
	// Database address
	DBADDRESS string
	// PrintVersion respresents flag to print program version and exit
	PrintVersion bool
	// ConsoleLogFlag respresents log level messages to the console stdout
	// To write to syslog this value should be 0
	ConsoleLogFlag = LevelDebug
)

func init() {
	if programName == "" {
		programName = "mbmi-go"
	}

	flag.StringVar(&ASSETSPATH, "A", "/usr/share/mbmi/assets", "Frontend")
	flag.StringVar(&SECRETPHRASE, "S", "", "Use static secret, othervise create it random on start")
	flag.StringVar(&SERVERADDRESS, "L", "127.0.0.1:8080", "Address listen on")
	flag.StringVar(&DBUSER, "Du", "nobody", "Database user")
	flag.StringVar(&DBPASS, "Dp", "", "Database user password")
	flag.StringVar(&DBNAME, "Db", "mail", "Database name")
	flag.StringVar(&DBADDRESS, "Dh", "localhost", "Database address")
	flag.IntVar(&ConsoleLogFlag, "v", 0, "Console verbose output, default 0 - off, 7 - debug")
	flag.BoolVar(&PrintVersion, "V", false, "Print version")
}

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

	// Create secret if empty
	if SECRETPHRASE == "" {
		if str, err := createSecret(32, false, false, false); err != nil {
			env.Fatal(err)
		} else {
			SECRETPHRASE = str
		}
	}

	if err := env.openDB(nil); err != nil {
		env.Fatal(err)
	}

	// Create router
	router = NewRouter()

	// Login
	router.Handle("POST", "/login", NewHandler(
		secretWrap(Login, SECRETPHRASE),
		env,
	))

	// Authentication tokens
	router.Handle("GET", "/application/jwt/:uid", NewHandler(
		Protect(GetUserJWT),
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
	router.Handle("PUT", "/alias/:aid", NewHandler(
		Protect(SetAlias),
		env,
	))
	router.Handle("DELETE", "/alias/:aid", NewHandler(
		Protect(DelAlias),
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
	router.Handle("POST", "/user", NewHandler(
		Protect(SetUser),
		env,
	))
	router.Handle("PUT", "/user/:uid", NewHandler(
		Protect(SetUser),
		env,
	))
	router.Handle("GET", "/password", NewHandler(
		Protect(Password),
		env,
	))

	// Accesses
	router.Handle("GET", "/accesses", NewHandler(
		Protect(Accesses),
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

	// Web hooks
	// Update imap logins
	router.Handle("POST", "/stat/imap/:uid", NewHandler(
		Protect(StatImapLogin),
		env,
	))

	// Services usage statists
	router.Handle("GET", "/servicestat", NewHandler(
		Protect(ServicesStat),
		env,
	))

	// Blind carbon copy (list)
	router.Handle("GET", "/bccs", NewHandler(
		Protect(Bccs),
		env,
	))

	// Blind carbon copy (item)
	router.Handle("GET", "/bcc/:bid", NewHandler(
		Protect(Bccs),
		env,
	))

	// Handle NotFound
	if ASSETSPATH != "" {
		router.NotFound = http.FileServer(http.Dir(ASSETSPATH))

		env.Notice("Using file server with public=%s for unknown routes", ASSETSPATH)
	}

	http.ListenAndServe(SERVERADDRESS, Middlewares(
		router,
		JWT(SECRETPHRASE, env),
		ApplicationToken(env),
		verbose(env),
		RequestId(),
	))
}

// Show program version
func showVersion(log LogIface) {
	var str = fmt.Sprintf("Mail boxes manager interface server (%s) %s, built %s", programName, programVersion, buildDate)

	if PrintVersion {
		fmt.Println(str)
		os.Exit(0)
	} else {
		if log != nil {
			log.Notice(str)
		}
	}
}
