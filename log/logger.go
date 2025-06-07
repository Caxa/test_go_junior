package log

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

type Config struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Init(config ...Config) {
	Logger = logrus.New()
	Logger.SetOutput(os.Stdout)

	// Установка уровня логирования
	logLevel := logrus.InfoLevel
	if len(config) > 0 && config[0].Level != "" {
		if level, err := logrus.ParseLevel(config[0].Level); err == nil {
			logLevel = level
		}
	} else if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		if level, err := logrus.ParseLevel(envLevel); err == nil {
			logLevel = level
		}
	}
	Logger.SetLevel(logLevel)

	// Установка формата
	if len(config) > 0 && config[0].Format == "json" {
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				path := strings.Split(f.File, "/")
				return "", fmt.Sprintf("%s:%d", path[len(path)-1], f.Line)
			},
		})
	} else {
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:          true,
			TimestampFormat:        "2006-01-02 15:04:05",
			ForceColors:            true,
			DisableLevelTruncation: true,
			PadLevelText:           true,
		})
	}

	// Добавляем информацию о вызывающем коде
	Logger.SetReportCaller(true)
}

// WithContext создает новую запись лога с контекстными полями
func WithContext(ctx context.Context) *logrus.Entry {
	fields := logrus.Fields{}

	// Добавляем поля из контекста
	if ctx != nil {
		if requestID := ctx.Value("request_id"); requestID != nil {
			fields["request_id"] = requestID
		}
		if userID := ctx.Value("user_id"); userID != nil {
			fields["user_id"] = userID
		}
	}

	// Добавляем информацию о месте вызова
	if pc, file, line, ok := runtime.Caller(1); ok {
		fName := runtime.FuncForPC(pc).Name()
		fields["file"] = file
		fields["line"] = line
		fields["func"] = fName[strings.LastIndex(fName, "/")+1:]
	}

	return Logger.WithFields(fields)
}
