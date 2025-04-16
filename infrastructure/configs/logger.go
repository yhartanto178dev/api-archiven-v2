package configs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.SugaredLogger

func InitializeLogger() {
	// Create logs directory if not exists
	if err := os.MkdirAll("logs", 0755); err != nil {
		panic(err)
	}
	// Get the current date in the desired format
	currentDate := time.Now().Format("02-01-2006") // Format: day-month-year
	logFileName := fmt.Sprintf("digital-archive-%s.log", currentDate)

	// Daily rotating file config
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join("logs", logFileName),
		MaxSize:    100, // megabytes
		MaxBackups: 30,  // keep 30 days
		MaxAge:     30,  // days
		Compress:   true,
		LocalTime:  true,
	})

	// Console output
	consoleWriter := zapcore.AddSync(os.Stdout)

	// Encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create core
	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			fileWriter,
			zap.InfoLevel,
		),
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			consoleWriter,
			zap.DebugLevel,
		),
	)

	// Create logger
	logger := zap.New(core, zap.AddCaller())
	Logger = logger.Sugar()
}

func SyncLogger() {
	_ = Logger.Sync()
}
