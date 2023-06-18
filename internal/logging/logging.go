package logging

import (
	"io"
	"log"
)

var Log = log.New(io.Discard, "DEBUG ", log.Ldate|log.Lmicroseconds)
