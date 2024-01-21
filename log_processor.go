package asynclog

import "fmt"

// processLogs is the method that processes log messages.
// This method runs in its own goroutine and handles messages sent to the LogChannel.
func (l *Logger) processLogs() {
	for logMessage := range l.LogChannel {
		if l.OutputToFile && logMessage.Level >= l.FileLevel {
			if logMessage.File == "" {
				logMessage.File = l.DefaultFileName
			}
			l.writeFile(logMessage.File, logMessage.FileMessage)
		}
		if l.OutputToConsole && logMessage.Level >= l.ConsoleLevel {
			fmt.Println(logMessage.ConsoleMessage)
		}
	}
}
