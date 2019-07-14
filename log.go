package prosumer

import (
	"os"

	"github.com/google/logger"
)

var (
	log *logger.Logger
)

func init() {
	lf := os.Stdout // default log to stdout
	logFile, ok := os.LookupEnv("PROSUMER_LOGFILE")
	if ok {
		var err error
		lf, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		if err != nil {
			logger.Fatalf("Failed to open log file: %v", err)
		}
	}
	log = logger.Init("prosumer", false, false, lf)

}
