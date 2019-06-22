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

func StrTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func In(a interface{}, arr ...interface{}) bool {
	for _, x := range arr {
		// https://stackoverflow.com/questions/34245932/checking-equality-of-interface
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

func Str(v interface{}) string {
	return fmt.Sprintf("%+v", v)
}

func StrArr(vs ...interface{}) []string {
	rv := []string{}
	for _, v := range vs {
		rv = append(rv, Str(v))
	}
	return rv
}

func ReprArr(arr ...interface{}) string {
	return fmt.Sprintf("[%s]", strings.Join(StrArr(arr...), ", "))
}

func ReprStrArr(arr ...string) string {
	return fmt.Sprintf("[%s]", strings.Join(arr, ", "))
}

func Sha256(content []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(content))
}
