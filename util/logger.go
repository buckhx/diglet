package util

import (
	"fmt"
	"log"
)

func Info(format string, vals ...interface{}) {
	log.Printf("INFO: "+format, vals...)
}

func Debug(format string, vals ...interface{}) {
	log.Printf("DEBUG: "+format, vals...)
}

func Warn(err error, extra string) {
	if err != nil {
		log.Printf("WARN: %s - %s", err, extra)
	}
}

func Error(err ...error) {
	log.Printf("ERROR: %s", err)
}

func Check(err error) {
	if err != nil {
		log.Fatal("FATAL ERROR: %s", err)
	}
}

func Checks(errs ...error) {
	for _, err := range errs {
		Check(err)
	}
}

func Sprintf(format string, vals ...interface{}) string {
	return fmt.Sprintf(format, vals...)
}

func Errorf(format string, vals ...interface{}) error {
	return fmt.Errorf(format, vals...)
}
