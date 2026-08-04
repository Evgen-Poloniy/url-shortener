package logs

import (
	"github.com/sirupsen/logrus"
)

func NewLogrusLogger(opts ...Option) *logrus.Logger {
	logger := logrus.New()

	options := &Options{
		level:  InfoLevel,
		format: JsonFormat,
	}

	for _, opt := range opts {
		opt(options)
	}

	// Set logger level.
	switch options.level {
	case TraceLevel:
		logger.SetLevel(logrus.TraceLevel)
	case DebugLevel:
		logger.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		logger.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		logger.SetLevel(logrus.ErrorLevel)
	case PanicLevel:
		logger.SetLevel(logrus.PanicLevel)
	case FatalLevel:
		logger.SetLevel(logrus.FatalLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// Set logger format.
	switch options.format {
	case TextFormat:
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	return logger
}
