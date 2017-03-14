package server

import (
	"log"
	"strconv"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func atoi64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 0)
}
