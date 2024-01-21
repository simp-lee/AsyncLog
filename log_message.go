package asynclog

// LogMessage represents a log message with its level, content, and additional parameters.
type LogMessage struct {
	Level          LogLevel               // Log level of the message (e.g., DEBUG, INFO, etc.)
	Message        string                 // The actual log message
	FileMessage    string                 // Formatted message for file output
	ConsoleMessage string                 // Formatted message for console output
	File           string                 // The target log file
	Params         map[string]interface{} // Additional parameters for the log message
}

// LogOption defines a function type for log message configuration.
type LogOption func(*LogMessage)

// SetLogFile specifies the log file for a log message.
// This function is used to set a custom log file for individual log messages.
func SetLogFile(file string) LogOption {
	return func(m *LogMessage) {
		m.File = file
	}
}

// SetLogParams specifies additional parameters for a log message.
// This function allows adding key-value pairs that provide additional information for the log message.
func SetLogParams(params map[string]interface{}) LogOption {
	return func(m *LogMessage) {
		m.Params = params
	}
}
