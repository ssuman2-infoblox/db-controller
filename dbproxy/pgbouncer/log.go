package pgbouncer

import (
	"strings"

	logimpl "github.com/sirupsen/logrus"
)

type pgbouncerLogger struct {
	log logger
}

type logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

var (
	// blogger is the logger used by the pgbouncerLogger.
	blogger = logimpl.StandardLogger()
)

func (l *pgbouncerLogger) Write(p []byte) (n int, err error) {
	// use a specific logger if one is set, useful for testing
	var log logger
	if l.log != nil {
		log = l.log
	} else {
		log = blogger
	}

	lines := strings.Split(string(p), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		parts := strings.SplitN(line, "]", 2)
		if len(parts) != 2 {
			log.Info(line)
			continue
		}
		line = strings.TrimSpace(parts[1])
		parts = strings.SplitN(line, " ", 2)
		switch parts[0] {
		case "DEBUG":
			log.Debug(parts[1])
		case "LOG":
			log.Info(parts[1])
		case "INFO":
			log.Info(parts[1])
		case "WARNING":
			log.Warn(parts[1])
		case "ERROR":
			log.Error(parts[1])
		case "FATAL":
			// don't actually exit
			log.Error(parts[1])
			log.Error("pgbouncer exited with FATAL error")
		default:
			log.Info(line)
		}
	}
	return len(p), nil
}
