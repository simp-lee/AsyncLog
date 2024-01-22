# AsyncLog: An Asynchronous Logging Package for Go

`AsyncLog` is a versatile and efficient asynchronous logging library for Go, designed for multi-level logging with support for custom formatting, colored output, and file logging capabilities.

## Features

- **Asynchronous Logging:** Processes log messages in a separate goroutine for minimal impact on main application flow.
- **Multiple Log Levels:** Supports levels like Trace, Debug, Info, Warning, Error, and Fatal for detailed logging.
- **Customizable Output:** Route log messages to different files or the console.
- **Colored console output:** Enhances readability with color-coded logs in the console.
- **Source Code Information:** Option to include source file and line number in logs.
- **Flexible Configuration:** Tailor logger behavior with functional options.
- **Parameter Formatting:** Choose between JSON or Key-Value formatting for log parameters.
- **File Logging:** Direct logs to files with configurable file names and output settings.

## Installation

Install AsyncLog using `go get`:

```bash
go get github.com/simp-lee/asynclog
```

## Quick Start

Here's a basic example to get started with `asynclog`:

```go
package main

import (
    "github.com/simp-lee/asynclog"
)

func main() {
    // Create a new logger with a buffer size of 100
    logger, err := asynclog.NewLogger(100)
    if err != nil {
        panic(err)
    }
    defer logger.Close() // Ensure logger is closed at the end
	
    // Logging at different levels
    logger.Trace("This is a trace message")
    logger.Debug("Debugging information here")
    logger.Info("Informational message")
    logger.Warning("This is a warning")
    logger.Error("Encountered an error")
    // Use Fatal sparingly - high severity
    logger.Fatal("Fatal error occurred")
}
```

## Configuration

Customize the logger at instantiation with various options:

```go
logger, err := asynclog.NewLogger(
    100,                                                     // Buffer size
    asynclog.SetFileLevel(asynclog.LogLevelInfo),            // Set file logging level
    asynclog.SetConsoleLevel(asynclog.LogLevelDebug),        // Set console logging level
    asynclog.EnableSourceInfo(true),                         // Enable source file information recording
    asynclog.SetDefaultFileName("app.log"),                  // Set default log file name
    asynclog.EnableFileOutput(false),                        // Disable file output
    asynclog.EnableConsoleOutput(true),                      // Enable console output
    asynclog.SetParamFormatter(asynclog.FormatParamsAsJSON), // Log parameter formatting
    asynclog.SetMaxFileHandles(20),                          // Set maximum number of file handles
)
```

## Parameters and Formatting

Include additional parameters in your log messages and customize their formatting style:

```go
params := map[string]interface{}{
    "user_id": 123,
    "action": "login",
}
logger.Info("User action", asynclog.SetLogParams(params)) // Log with additional parameters
```

## Contributing

Your contributions to `AsyncLog` are welcome! Feel free to open issues or submit pull requests for improvements or new features.