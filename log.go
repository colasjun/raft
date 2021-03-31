package main

import (
	"log"
	"os"
)

func createLogger() *log.Logger {
	fname := "./raft.log"
	file, _ := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	return log.New(file, "raft log: ", log.Lshortfile);
}
