package asynclog

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"strings"
)

// formatLogLevel formats the log level string with optional color and bold styling.
func formatLogLevel(text string, level LogLevel, bold bool) string {
	colorAttr := getColorAttribute(level)
	formatter := color.New(colorAttr)
	if bold {
		formatter = formatter.Add(color.Bold)
	}
	return formatter.SprintfFunc()(text)
}

// getColorAttribute returns the color attribute based on the log level.
func getColorAttribute(level LogLevel) color.Attribute {
	switch level {
	case LogLevelTrace:
		return color.FgCyan
	case LogLevelDebug:
		return color.FgMagenta
	case LogLevelInfo:
		return color.FgGreen
	case LogLevelWarning:
		return color.FgYellow
	case LogLevelError:
		return color.FgRed
	case LogLevelFatal:
		return color.FgHiRed
	default:
		return color.Faint
	}
}

// formatParamsWithColor formats the additional parameters with a lighter color.
func formatParamsWithColor(params string) string {
	if params == "" {
		return ""
	}
	return color.New(color.Faint).SprintfFunc()(params)
}

// 返回代表日志级别的字符串
func (level LogLevel) String() string {
	switch level {
	case LogLevelTrace:
		return "TRACE"
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarning:
		return "WARNING"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// FormatParamsAsKeyValue formats parameters as key-value pairs.
func FormatParamsAsKeyValue(params map[string]interface{}) string {
	if len(params) == 0 {
		return "" // Return empty string if no parameters
	}
	var builder strings.Builder
	for key, value := range params {
		builder.WriteString(fmt.Sprintf("  \"%s\": %v\n", key, value))
	}

	return strings.TrimSuffix(builder.String(), "\n")
}

// FormatParamsAsJSON formats parameters as a JSON string.
func FormatParamsAsJSON(params map[string]interface{}) string {
	if len(params) == 0 {
		return "" // Return empty string if no parameters
	}
	jsonBytes, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting params: %v", err)
	}
	return string(jsonBytes)
}
