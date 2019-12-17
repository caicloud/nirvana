/*
Copyright 2019 Caicloud Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

// goLogger is the logger with go logger.
type goLogger struct {
	l         *log.Logger
	level     Level
	calldepth int
}

// NewGoStandardLogger creates a go logger.
func NewGoStandardLogger(level Level, out io.Writer) Logger {
	return &goLogger{
		l:         log.New(out, "", 0),
		level:     level,
		calldepth: 2,
	}
}

// V reports whether verbosity at the call site is at least the requested level.
// The returned value is a Verboser, which implements Info, Infof
// and Infoln. These methods will write to the Info log if called.
func (l *goLogger) V(v Level) Verboser {
	if v > l.level {
		return silentLogger
	}
	return l
}

// Info logs to the INFO log.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (l *goLogger) Info(a ...interface{}) {
	l.output(SeverityInfo, a...)
}

// Infof logs to the INFO log.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func (l *goLogger) Infof(format string, a ...interface{}) {
	l.outputf(SeverityInfo, format, a...)
}

// Infoln logs to the INFO log.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (l *goLogger) Infoln(a ...interface{}) {
	l.outputln(SeverityInfo, a...)
}

// Warning logs to the WARNING logs.
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func (l *goLogger) Warning(a ...interface{}) {
	l.output(SeverityWarning, a...)
}

// Warningf logs to the WARNING logs.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func (l *goLogger) Warningf(format string, a ...interface{}) {
	l.outputf(SeverityWarning, format, a...)
}

// Warningln logs to the WARNING logs.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (l *goLogger) Warningln(a ...interface{}) {
	l.outputln(SeverityWarning, a...)
}

// Error logs to the ERROR logs.
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func (l *goLogger) Error(a ...interface{}) {
	l.output(SeverityError, a...)
}

// Errorf logs to the ERROR logs.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func (l *goLogger) Errorf(format string, a ...interface{}) {
	l.outputf(SeverityError, format, a...)
}

// Errorln logs to the ERROR logs.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (l *goLogger) Errorln(a ...interface{}) {
	l.outputln(SeverityError, a...)
}

// Fatal logs to the FATAL logs, then calls os.Exit(1).
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func (l *goLogger) Fatal(a ...interface{}) {
	l.output(SeverityFatal, a...)
}

// Fatalf logs to the FATAL logs, then calls os.Exit(1).
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func (l *goLogger) Fatalf(format string, a ...interface{}) {
	l.outputf(SeverityFatal, format, a...)
}

// Fatalln logs to the FATAL logs, then calls os.Exit(1).
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (l *goLogger) Fatalln(a ...interface{}) {
	l.outputln(SeverityFatal, a...)
}

// Clone clones current logger with new wrapper.
// A positive wrapper indicates how many wrappers outside the logger.
func (l *goLogger) Clone(wrapper int) Logger {
	return &goLogger{
		level:     l.level,
		l:         l.l,
		calldepth: l.calldepth + wrapper,
	}
}

func (l *goLogger) output(severity Severity, a ...interface{}) {
	// Set nolint, the potential error is not handled in go
	// standard log package, too.
	l.l.Output(l.calldepth, prefix(severity, 2)+fmt.Sprint(a...)) //nolint
	l.exitIfFatal(severity)
}

func (l *goLogger) outputf(severity Severity, format string, a ...interface{}) {
	// Set nolint, the potential error is not handled in go
	// standard log package, too.
	l.l.Output(l.calldepth, prefix(severity, 2)+fmt.Sprintf(format, a...)) //nolint
	l.exitIfFatal(severity)
}

func (l *goLogger) outputln(severity Severity, a ...interface{}) {
	// Set nolint, the potential error is not handled in go
	// standard log package, too.
	l.l.Output(l.calldepth, prefix(severity, 2)+fmt.Sprintln(a...)) //nolint
	l.exitIfFatal(severity)
}

func (l *goLogger) exitIfFatal(severity Severity) {
	if severity == SeverityFatal {
		os.Exit(1)
	}
}
