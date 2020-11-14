package steno

import (
	"fmt"
	"os"
	"path/filepath"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

type NullLogger struct{}

func (l NullLogger) Printf(format string, v ...interface{}) {
}

type FileLog struct {
	log *os.File
}

func NewLog(logFile string) (*FileLog, error) {

	d := &FileLog{}

	err := os.MkdirAll(filepath.Dir(logFile), os.ModePerm)
	if err != nil {
		return nil, err
	}

	f, err := os.Create(logFile)
	if err == nil {
		d.log = f
	} else {
		return nil, err
	}

	return d, nil
}

func (d *FileLog) Close() {
	if d.log != nil {
		d.log.Close()
		d.log = nil
	}
}

func (d *FileLog) Printf(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format, v...)
	if d.log != nil {
		fmt.Fprintf(d.log, format, v...)
		d.log.Sync()
	}
}

/*
func (d *FileLog) Println(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	if d.log != nil {
		fmt.Fprintln(d.log, v...)
		d.log.Sync()
	}
}
*/
