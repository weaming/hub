package core

import (
	"crypto/sha256"
	"fmt"
	"log"
	"strings"
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

func PrettyTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func In(a interface{}, arr ...interface{}) bool {
	for _, x := range arr {
		// TODO https://stackoverflow.com/questions/34245932/checking-equality-of-interface
		if a == x {
			return true
		}
	}
	return false
}

func InStrArr(a string, arr ...string) bool {
	for _, x := range arr {
		if a == x {
			return true
		}
	}
	return false
}

func StrArr2Str(arr []string) string {
	return fmt.Sprintf("[%s]", strings.Join(arr, ", "))
}

func Sha256(content []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(content))
}
