package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	// Program name
	NAME,
	// Program version
	VERSION,
	// Listen
	SERVERADDRESS,
	// Assets
	ASSETSPATH,
	// Build date and time
	BUILDDATE,
	// Database user
	DBUSER,
	// Database user password
	DBPASS,
	// Database name
	DBNAME,
	// Database address
	DBADDRESS string
	// Пиши сообщения в консоль
	CONSOLELOG = LevelDebug

	PrintVersion bool
)

func init() {
	if NAME == "" {
		NAME = "mbmi-go"
	}

	flag.StringVar(&ASSETSPATH, "A", "/usr/share/mbmi/assets", "Frontend")
	flag.StringVar(&SERVERADDRESS, "L", "127.0.0.1:8080", "Address listen on")
	flag.StringVar(&DBUSER, "Du", "nobody", "Database user")
	flag.StringVar(&DBPASS, "Dp", "", "Database user password")
	flag.StringVar(&DBNAME, "Db", "mail", "Database name")
	flag.StringVar(&DBADDRESS, "Dh", "localhost", "Database address")
	flag.IntVar(&CONSOLELOG, "v", 0, "Console verbose output, default 0 - off, 7 - debug")
	flag.BoolVar(&PrintVersion, "V", false, "Print version")
}

// Show program version
func showVersion(log LogIface) {
	var str = fmt.Sprintf("Mail boxes manager interface server (%s) %s, built %s", NAME, VERSION, BUILDDATE)

	if PrintVersion {
		fmt.Println(str)
		os.Exit(0)
	} else {
		if log != nil {
			log.Notice(str)
		}
	}
}
