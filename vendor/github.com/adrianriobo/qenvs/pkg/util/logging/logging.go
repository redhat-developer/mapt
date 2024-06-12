package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	logfile       *os.File
	LogLevel      string
	originalHooks = logrus.LevelHooks{}
)

func OpenLogFile(basePath string, fileName string) (*os.File, error) {
	logFile, err := os.OpenFile(
		filepath.Join(basePath, fileName),
		os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func CloseLogging() {
	logfile.Close()
	logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))
}

func BackupLogFile() {
	if logfile == nil {
		return
	}
	os.Rename(logfile.Name(), fmt.Sprintf("%s_%s", logfile.Name(), time.Now().Format("20060102150405"))) // nolint
}

func InitLogrus(basePath, fileName string) {
	// var err error
	// logfile, err = OpenLogFile(basePath, fileName)
	// if err != nil {
	// 	logrus.Fatal("Unable to open log file: ", err)
	// }
	// send logs to file and console
	// logrus.SetOutput(io.MultiWriter(logfile, os.Stdout))
	logrus.SetOutput(io.MultiWriter(os.Stdout))
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	for k, v := range logrus.StandardLogger().Hooks {
		originalHooks[k] = v
	}

}

func GetWritter() *io.PipeWriter {
	return logrus.StandardLogger().Writer()
}

func Info(args ...interface{}) {
	logrus.Info(args...)
}

func Infof(s string, args ...interface{}) {
	logrus.Infof(s, args...)
}

// func InfofWithFields(s string, args ...interface{}) {
// 	logrus.Fields
// 	logrus.WithFields()(s, args...)
// }

// log.WithFields(logrus.Fields{
//     "animal": "walrus",
//     "size":   10,
//   }).Info("A group of walrus emerges from the ocean")

func Warn(args ...interface{}) {
	logrus.Warn(args...)
}

func Warnf(s string, args ...interface{}) {
	logrus.Warnf(s, args...)
}

func Fatal(args ...interface{}) {
	logrus.Fatal(args...)
}

func Fatalf(s string, args ...interface{}) {
	logrus.Fatalf(s, args...)
}

func Error(args ...interface{}) {
	logrus.Error(args...)
}

func Errorf(s string, args ...interface{}) {
	logrus.Errorf(s, args...)
}

func Debug(args ...interface{}) {
	logrus.Debug(args...)
}

func Debugf(s string, args ...interface{}) {
	logrus.Debugf(s, args...)
}
