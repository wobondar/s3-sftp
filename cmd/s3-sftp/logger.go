package main

import (
	"log"
	"os"
)

// Logger setup
func setupAppLogger(a *App) {
	a.log.info = log.New(os.Stdout, "INF: ", log.Ldate|log.Ltime|log.Lmicroseconds)
	a.log.err = log.New(os.Stderr, "ERR: ", log.Ldate|log.Ltime|log.Lmicroseconds)
	if a.log.debugMode {
		a.log.debug = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lmicroseconds)
		a.log.debug.Println("Debug mode enabled")
	}
}

// debug logs a message if debug mode is enabled
func (a *App) debug(format string, v ...interface{}) {
	if a.log.debugMode && a.log.debug != nil {
		a.log.debug.Printf(format, v...)
	}
}
