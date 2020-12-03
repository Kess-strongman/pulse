package utils

import (
	"fmt"
	"log"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Logging makes a call to start rolling logging and sets the prefixes and time and date flags
func Logging() {
	fmt.Println("Initialising Logging")
	rollingLog()
	log.SetPrefix("PLS: ")                       // All messages will be prefixed by OWS:
	log.SetFlags(log.LstdFlags | log.Lshortfile) // Time, date,

}

func rollingLog() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "/var/log/pulse.log",
		MaxSize:    5, // megabytes
		MaxBackups: 30,
		MaxAge:     28,    //days
		Compress:   false, // disabled by default
	})
}
