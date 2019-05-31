package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

// The hidden and user-facing loggers.
var (
	logFile = logrus.New()
	logGui  = logrus.New()
)

// redirectStderr attempts to redirect stderr to abv.log. If this is not
// possible, error messages are instead directed to the default stderr
// destination.
func redirectStderr(logger *logrus.Logger) (file *os.File) {
	file, err := os.OpenFile(conf.GetString("configPath")+"/abv.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger.Out = file
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}
	return file
}

// logAllDebug logs a debug message to both the gui and log file.
func logAllDebug(args ...interface{}) {
	logGui.Debug(args...)
	logFile.Debug(args...)
}

// logAllError logs an error message to both the gui and log file.
func logAllError(args ...interface{}) {
	logGui.Error(args...)
	logFile.Error(args...)
}

// logAllInfo logs an info message to both the gui and log file.
func logAllInfo(args ...interface{}) {
	logGui.Info(args...)
	logFile.Info(args...)
}

// logAllWarn logs a user warning to both the gui and log file.
func logAllWarn(args ...interface{}) {
	logGui.Warn(args...)
	logFile.Warn(args...)
}
