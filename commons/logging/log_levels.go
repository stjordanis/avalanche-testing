package logging

import (
	"github.com/sirupsen/logrus"
)

const (
	trace = "trace"
	debug = "debug"
	info = "info"
	warn = "warn"
	error = "error"
	fatal = "fatal"
)

var acceptableLogLevels = map[string]logrus.Level{
	trace: logrus.TraceLevel,
	debug: logrus.DebugLevel,
	info: logrus.InfoLevel,
	warn: logrus.WarnLevel,
	error: logrus.ErrorLevel,
	fatal: logrus.FatalLevel,
}

func LevelFromString(str string) *logrus.Level {
	level, found := acceptableLogLevels[str]
	if !found {
		return nil
	}
	return &level
}

func GetAcceptableStrings() []string {
	return []string {
		trace,
		debug,
		info,
		warn,
		error,
		fatal,
	}
}
