package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

var (
	logFile = logrus.New()
	logGui  = logrus.New()
)

func init() {
	quantity = 1

	//Setup loggers
	f := logrus.TextFormatter{}
	f.ForceColors = true
	f.DisableTimestamp = true
	f.DisableLevelTruncation = true
	logGui.Formatter = &f
	logGui.SetLevel(logrus.InfoLevel)
	logFile.SetLevel(logrus.DebugLevel)
}

func redirectStderr(logger *logrus.Logger) (file *os.File) {
	file, err := os.OpenFile(conf.GetString("configPath") + "/abv.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger.Out = file
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}
	return file
}

func logAllDebug(args ...interface{}) {
	logGui.Debug(args...)
	logFile.Debug(args...)
}

func logAllError(args ...interface{}) {
	logGui.Error(args...)
	logFile.Error(args...)
}

func logAllInfo(args ...interface{}) {
	logGui.Info(args...)
	logFile.Info(args...)
}

func logAllWarn(args ...interface{}) {
	logGui.Warn(args...)
	logFile.Warn(args...)
}
