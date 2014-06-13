package log

import (
	"strings"

	"github.com/wsxiaoys/terminal/color"
)

/*
Log is the interface for all general logging methods.
*/
type Log interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
	Panicf(format string, v ...interface{})
	Panicln(v ...interface{})
}

/*
Logger is almost identical to Log except that it also contains the Write(p
[]byte) method so it can satisfy tye io.Writer interface.  At some point, these
two types should probably be combined.
*/
type Logger interface {
	Log
	Write(p []byte) (int, error)
}

/*
An OutWriter is responsible for for implementing the io.Writer interface.
*/
type OutWriter struct {
	Log
	fmtString string
}

/*
NewOutWriter accepts a logger and a format string and returns an OutWriter.
When written to, the OutWriter will take the input, split it into lines, and
print it to the logger using the provided format string.  The intended use case
of this functionality is for printing nice, colorful messages
*/
func NewOutWriter(logger Log, fmtString string) *OutWriter {
	return &OutWriter{
		Log:       logger,
		fmtString: fmtString,
	}
}

/*
Write writes the provided bytes, one line at a time, after interpolating them
into the provided format string, to the provided logger.
*/
func (ow *OutWriter) Write(p []byte) (n int, err error) {
	lines := strings.Split(string(p), "\n")
	for _, line := range lines {
		ow.Print(color.Sprintf(ow.fmtString, line))
	}

	return len(p), nil
}
