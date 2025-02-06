package utils

import (
	"io"
	"log"
)

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func CloseAndLogError(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Println(err)
	}
}

func LogError(err error) {
	if err != nil {
		log.Println("Error:", err)
	}
}
