package util

import (
	"fmt"
	"log"
)

var DEBUG = false

func Info(format string, vals ...interface{}) {
	log.Printf("INFO: "+format, vals...)
}

func Debug(format string, vals ...interface{}) {
	if DEBUG {
		log.Printf("DEBUG: "+format, vals...)
	}
}

func Warn(err error, extra string) {
	if err != nil {
		log.Printf("WARN: %s - %s", err, extra)
	}
}

func Error(err ...error) {
	log.Printf("ERROR: %s", err)
}

func Fatal(format string, vals ...interface{}) {
	msg := Sprintf("Diglet fainted! "+format, vals...)
	log.Fatal(msg)
}

func Check(err error) {
	if err != nil {
		panic(err)
		Fatal("%s", err)
	}
}

func Checks(errs ...error) {
	for _, err := range errs {
		Check(err)
	}
}

func Printf(format string, vals ...interface{}) {
	fmt.Printf(format, vals...)
}

func Println(s string) {
	fmt.Println(s)
}

func Sprintf(format string, vals ...interface{}) string {
	return fmt.Sprintf(format, vals...)
}

func Errorf(format string, vals ...interface{}) error {
	return fmt.Errorf(format, vals...)
}
