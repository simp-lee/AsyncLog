package asynclog

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel defines the severity of a log message.
type LogLevel int

const (
	LogLevelTrace LogLevel = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarning
	LogLevelError
	LogLevelFatal

	// DefaultBufferSize is the default size of the log message channel.
	DefaultBufferSize = 100

	// DefaultFileName is the default name for log files.
	DefaultFileName = "default.log"

	// DefaultMaxFileHandles is the default maximum number of file handles.
	DefaultMaxFileHandles = 10

	// DefaultUnusedFileHandleThreshold is the threshold for cleaning up unused file handles.
	DefaultUnusedFileHandleThreshold = 30 * time.Minute

	// DefaultCleanupTicker is the interval for the cleanup routine.
	DefaultCleanupTicker = 10 * time.Minute
)

// ParamFormatter is a function type for formatting log parameters.
type ParamFormatter func(map[string]interface{}) string

// Logger represents an asynchronous logger.
type Logger struct {
	LogChannel      chan LogMessage      // Channel for log messages.
	FileLevel       LogLevel             // Minimum level of messages to log to file.
	ConsoleLevel    LogLevel             // Minimum level of messages to log to console.
	DefaultFileName string               // Default log file name.
	OutputToFile    bool                 // Flag to enable or disable file output.
	OutputToConsole bool                 // Flag to enable or disable console output.
	paramFormatter  ParamFormatter       // Function used to format log parameters.
	fileHandles     map[string]*os.File  // File handles for each log file.
	fileAccessTimes map[string]time.Time // Last access time for each file handle.
	fileMutex       sync.Mutex           // Mutex for synchronizing file access.
	maxFileHandles  int                  // Maximum number of file handles.
	AddSource       bool                 // Flag to add source file info in logs.
}

// LoggerOption defines a function type for logger configuration options.
type LoggerOption func(*Logger) error

// NewLogger creates a new Logger with specified options.
// opts are functional options to configure the Logger.
func NewLogger(opts ...LoggerOption) (*Logger, error) {
	logger := &Logger{
		LogChannel:      make(chan LogMessage, DefaultBufferSize), // Default size of the log message channel
		FileLevel:       LogLevelInfo,                             // Default file logging level.
		ConsoleLevel:    LogLevelDebug,                            // Default console logging level.
		DefaultFileName: DefaultFileName,                          // Default file name for logging.
		OutputToFile:    true,                                     // Enable logging to file by default.
		OutputToConsole: true,                                     // Enable logging to console by default.
		paramFormatter:  FormatParamsAsKeyValue,                   // Default parameter formatter set to KeyValue.
		AddSource:       false,                                    // Source file info is disabled by default.
		fileHandles:     make(map[string]*os.File),
		fileAccessTimes: make(map[string]time.Time),
		maxFileHandles:  DefaultMaxFileHandles,
	}

	// Apply each configuration option to the logger
	for _, opt := range opts {
		if err := opt(logger); err != nil {
			return nil, err
		}
	}

	// Start the cleanup ticker routine.
	go func() {
		cleanupTicker := time.NewTicker(DefaultCleanupTicker)
		for {
			select {
			case <-cleanupTicker.C:
				logger.cleanupUnusedFileHandles()
			}
		}
	}()

	// Start the log processing goroutine
	go logger.processLogs()

	return logger, nil
}

// SetBufferSize sets the size of the log message channel.
func SetBufferSize(size int) LoggerOption {
	return func(l *Logger) error {
		if size <= 0 {
			return fmt.Errorf("bufferSize must be positive")
		}
		l.LogChannel = make(chan LogMessage, size)
		return nil
	}
}

// SetFileLevel sets the file log level.
func SetFileLevel(level LogLevel) LoggerOption {
	return func(l *Logger) error {
		l.FileLevel = level
		return nil
	}
}

// SetConsoleLevel sets the console log level.
func SetConsoleLevel(level LogLevel) LoggerOption {
	return func(l *Logger) error {
		l.ConsoleLevel = level
		return nil
	}
}

// EnableSourceInfo enables or disables the logging of source file information.
func EnableSourceInfo(enable bool) LoggerOption {
	return func(l *Logger) error {
		l.AddSource = enable
		return nil
	}
}

// SetDefaultFileName sets the default log file name.
func SetDefaultFileName(fileName string) LoggerOption {
	return func(l *Logger) error {
		l.DefaultFileName = fileName
		return nil
	}
}

// EnableFileOutput enables or disables file output.
func EnableFileOutput(enable bool) LoggerOption {
	return func(l *Logger) error {
		l.OutputToFile = enable
		return nil
	}
}

// EnableConsoleOutput enables or disables console output.
func EnableConsoleOutput(enable bool) LoggerOption {
	return func(l *Logger) error {
		l.OutputToConsole = enable
		return nil
	}
}

// SetParamFormatter sets the parameter formatting strategy for the logger.
func SetParamFormatter(formatter ParamFormatter) LoggerOption {
	return func(l *Logger) error {
		l.paramFormatter = formatter
		return nil
	}
}

// SetMaxFileHandles sets the maximum number of file handles.
func SetMaxFileHandles(maxHandles int) LoggerOption {
	return func(l *Logger) error {
		if maxHandles <= 0 {
			return fmt.Errorf("maxFileHandles must be positive")
		}
		l.maxFileHandles = maxHandles
		return nil
	}
}

func (l *Logger) Close() {
	l.fileMutex.Lock()
	defer l.fileMutex.Unlock()

	for _, file := range l.fileHandles {
		if err := file.Close(); err != nil {
			fmt.Printf("Failed to close log file: %v", err)
		}
	}
}

// log is an internal method to log a message with given options.
// It formats the message based on the log level, and sends it to the LogChannel.
// This method is used by public methods like Debug, Info, Warning, Error.
func (l *Logger) log(level LogLevel, message string, opts ...LogOption) {
	// If the log level is not sufficient for file or console output, skip processing
	if level < l.FileLevel && level < l.ConsoleLevel {
		return
	}

	// Prepare the log message
	logMsg := LogMessage{
		Level:   level,
		Message: message,
		File:    l.DefaultFileName, // Default log file
		Params:  make(map[string]interface{}),
	}

	// Apply each option to the LogMessage
	for _, opt := range opts {
		opt(&logMsg)
	}

	// Format the current time
	timestamp := time.Now().Format("2006/01/02 15:04:05")

	// Format log parameters
	formattedParams := l.paramFormatter(logMsg.Params)

	var sourceInfo, fileMessage, consoleMessage string

	// Prepare source information
	if l.AddSource {
		callerFile, callerLine := getCallerInfo()
		sourceInfo = fmt.Sprintf("[%s:%d]", filepath.Base(callerFile), callerLine)
	}

	// Prepare the log message for file output
	if level >= l.FileLevel {
		fileMessage = l.prepareFileMessage(timestamp, sourceInfo, level, logMsg.Message, formattedParams)
	}

	// Prepare the log message for console output
	if level >= l.ConsoleLevel {
		consoleMessage = l.prepareConsoleMessage(timestamp, sourceInfo, level, logMsg.Message, formattedParams)
	}

	// Send the message to the LogChannel
	l.LogChannel <- LogMessage{
		Level:          level,
		FileMessage:    fileMessage,
		ConsoleMessage: consoleMessage,
		File:           logMsg.File,
	}
}

// prepareFileMessage formats the log message for file output.
func (l *Logger) prepareFileMessage(timestamp, sourceInfo string, level LogLevel, message, formattedParams string) string {
	fileMessage := fmt.Sprintf("[%s]%s %s: %s", timestamp, sourceInfo, level.String(), message)
	if formattedParams != "" {
		fileMessage += "\n" + formattedParams
	}
	return fileMessage
}

// prepareConsoleMessage formats the log message for console output with color.
func (l *Logger) prepareConsoleMessage(timestamp, sourceInfo string, level LogLevel, message, formattedParams string) string {
	coloredLevel := formatLogLevel(level.String(), level, true) // Colored and bold level
	coloredMessage := formatLogLevel(message, level, false)     // Colored message without bold
	consoleMessage := fmt.Sprintf("[%s]%s %s: %s", timestamp, sourceInfo, coloredLevel, coloredMessage)
	if formattedParams != "" {
		consoleMessage += "\n" + formatParamsWithColor(formattedParams)
	}
	return consoleMessage
}

// Trace logs a message at the Trace level.
func (l *Logger) Trace(message string, opts ...LogOption) {
	l.log(LogLevelTrace, message, opts...)
}

// Debug logs a message at the Debug level.
func (l *Logger) Debug(message string, opts ...LogOption) {
	l.log(LogLevelDebug, message, opts...)
}

// Info logs a message at the Info level.
func (l *Logger) Info(message string, opts ...LogOption) {
	l.log(LogLevelInfo, message, opts...)
}

// Warning logs a message at the Warning level.
func (l *Logger) Warning(message string, opts ...LogOption) {
	l.log(LogLevelWarning, message, opts...)
}

// Error logs a message at the Error level.
func (l *Logger) Error(message string, opts ...LogOption) {
	l.log(LogLevelError, message, opts...)
}

// Fatal logs a message at the Fatal level.
func (l *Logger) Fatal(message string, opts ...LogOption) {
	l.log(LogLevelFatal, message, opts...)
}
