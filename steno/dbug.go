package main

import (
	"fmt"
	"os"
	"time"
)

/*
type Logger interface {
	Printf(format string, v ...interface{})
}

type NullLogger struct{}

func (l NullLogger) Printf(format string, v ...interface{}) {
}
*/

type dbugLog struct {
	log *os.File
}

func NewDbugLog(logFile string) *dbugLog {

	d := &dbugLog{}
	f, err := os.Create(logFile)
	if err == nil {
		d.log = f
	} else {
		fmt.Fprintf(os.Stderr, "Can't open %s (%s) - running without log\n", logFile, err)
	}

	d.Printf("startup %s\n", time.Now().Format("2006-01-02 15:04:05"))
	return d
}

func (d *dbugLog) Close() {
	d.Printf("shutdown %s\n", time.Now().Format("2006-01-02 15:04:05"))
	if d.log != nil {
		d.log.Close()
		d.log = nil
	}
}

func (d *dbugLog) Printf(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format, v...)
	if d.log != nil {
		fmt.Fprintf(d.log, format, v...)
	}
}

func (d *dbugLog) Println(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	if d.log != nil {
		fmt.Fprintln(d.log, v...)
	}
}
