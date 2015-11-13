package burrow

import (
	"encoding/binary"
	"fmt"
	"log"
	"strconv"
	"strings"
)

func info(format string, vals ...interface{}) {
	log.Printf("Burrow info: "+format, vals...)

}

func warn(err error, extra string) {
	if err != nil {
		log.Printf("Burrow warning: %s - %s", err, extra)
	}
}

func errorlog(err ...error) {
	log.Printf("Burrow error: %s", err)
}

func check(err error) {
	if err != nil {
		log.Fatal("Fatal Burrow error: %s", err)
	}
}

func checks(errs ...error) {
	for _, err := range errs {
		check(err)
	}
}

func sprintSizeOf(v interface{}) string {
	return strconv.Itoa(binary.Size(v))
}

func sprintf(format string, vals ...interface{}) string {
	return fmt.Sprintf(format, vals...)
}

func errorf(format string, vals ...interface{}) error {
	return fmt.Errorf(format, vals...)
}

func itoa(v int) string {
	return strconv.Itoa(v)
}

func atoi(v string) (int, error) {
	return strconv.Atoi(v)
}

func atof(v string) (float64, error) {
	return strconv.ParseFloat(v, 64)
}

func toLower(v string) string {
	return strings.ToLower(v)
}

func assertString(v interface{}) (err error) {
	if _, ok := v.(string); !ok {
		err = fmt.Errorf("Cannot assert string %q", v)
	}
	return
}

func assertNumber(v interface{}) (err error) {
	if _, ok := v.(float64); !ok {
		err = fmt.Errorf("Cannot assert number %q", v)
	}
	return
}
func validateParams(params map[string]interface{}, keys []string) (err error) {
	for _, key := range keys {
		if _, ok := params[key]; !ok {
			err = fmt.Errorf("Missing param: %q", key)
		}
	}
	return
}
