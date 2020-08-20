package logger

import (

	"github.com/Crypta-Eve/spyglass2/config"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"github.com/snowzach/rotatefilehook"
	"time"
)

var Log = log.New()

// Configure will set up a global Log in a decent format and with the specified level.
// level should be one of: DEBUG, INFO, WARNING, ERROR, FATAL, PANIC
func Configure(level string) {
	formatter := &log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "02-01-2006 15:04:05",
	}
	Log.SetFormatter(formatter)

	switch level {
	case "DEBUG":
		Log.SetLevel(log.DebugLevel)
	case "INFO":
		Log.SetLevel(log.InfoLevel)
	case "WARNING":
		Log.SetLevel(log.WarnLevel)
	case "ERROR":
		Log.SetLevel(log.ErrorLevel)
	case "FATAL":
		Log.SetLevel(log.FatalLevel)
	case "PANIC":
		Log.SetLevel(log.PanicLevel)
	default:
		Log.SetLevel(log.InfoLevel)
		Log.WithFields(log.Fields{
			"requested_level": level,
			"set_level":       "INFO",
		}).Info("invalid log level, setting default")
	}

	Log.Info("Log configured")
}

func ConfigureLogFile() error {
	logPath := config.GetConfig().LogPath
	logName := "spyglass.log"
	fullpath := filepath.Join(logPath, logName)

	rfh, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
		Filename:   fullpath,
		MaxSize:    512,
		MaxBackups: 3,
		MaxAge:     60,
		Level:      Log.Level,
		Formatter:  &log.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC822,
		},
	})

	if err != nil {
		Log.WithField("err", err).Error("failed to init log file handler")
	}

	Log.AddHook(rfh)

	Log.Info("log file hook configured")

	return nil

}
