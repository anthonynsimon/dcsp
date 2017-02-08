package dcsp

import (
	"os"

	log "github.com/Sirupsen/logrus"
)

// TODO: make logging configurable
func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
}
