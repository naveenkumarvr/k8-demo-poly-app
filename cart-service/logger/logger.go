package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger initializes a Zap logger with production configuration
// Logs are written to both stdout and /var/log/app/cart-service.log
// This supports both Docker logging driver capture and sidecar log shipping
func InitLogger(serviceName, podName, nodeName, environment string) (*zap.Logger, error) {
	// Create encoder config for JSON format
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create JSON encoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Create log directory if it doesn't exist (for local development)
	logDir := "/var/log/app"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// If we can't create the directory, only log to stdout
		// This handles cases where we don't have permissions (e.g., in distroless)
		core := zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			zapcore.InfoLevel,
		)
		logger := zap.New(core, zap.AddCaller())
		logger.Warn("Failed to create log directory, logging to stdout only", zap.Error(err))

		// Add service metadata fields
		return logger.With(
			zap.String("service", serviceName),
			zap.String("pod_name", podName),
			zap.String("node_name", nodeName),
			zap.String("environment", environment),
		), nil
	}

	// Open log file for writing
	logFile, err := os.OpenFile(logDir+"/cart-service.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// If we can't open the file, only log to stdout
		core := zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			zapcore.InfoLevel,
		)
		logger := zap.New(core, zap.AddCaller())
		logger.Warn("Failed to open log file, logging to stdout only", zap.Error(err))

		return logger.With(
			zap.String("service", serviceName),
			zap.String("pod_name", podName),
			zap.String("node_name", nodeName),
			zap.String("environment", environment),
		), nil
	}

	// Create multi-writer to write to both stdout and file
	// This allows Docker to capture logs via stdout while sidecar reads from file
	multiWriter := zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(os.Stdout),
		zapcore.AddSync(logFile),
	)

	// Create core with both outputs
	core := zapcore.NewCore(
		encoder,
		multiWriter,
		zapcore.InfoLevel,
	)

	// Create logger with caller information
	logger := zap.New(core, zap.AddCaller())

	// Add service metadata fields that will appear in every log entry
	return logger.With(
		zap.String("service", serviceName),
		zap.String("pod_name", podName),
		zap.String("node_name", nodeName),
		zap.String("environment", environment),
	), nil
}
