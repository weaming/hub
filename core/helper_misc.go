package core

import (
	"log"
	"time"
)

func FatalErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func PrintErr(err error) bool {
	if err != nil {
		log.Println(err)
	}
	return err != nil
}

func prettyTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
