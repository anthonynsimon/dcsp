package dcsp

import (
	"os"

	"io"

	log "github.com/Sirupsen/logrus"
)

func init() {
	SetLogFormatter(&log.TextFormatter{})
	SetLogOutput(os.Stdout)
	SetLogLevel(log.DebugLevel)
}

func SetLogFormatter(formatter log.Formatter) {
	log.SetFormatter(formatter)
}

func SetLogOutput(out io.Writer) {
	log.SetOutput(out)
}

func SetLogLevel(level log.Level) {
	log.SetLevel(level)
}
