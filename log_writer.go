package asynclog

import (
	"fmt"
	"os"
	"time"
)

// writeFile writes a message to the specified file.
// It's responsible for opening and maintaining file handles,
// as well as writing log messages to these files.
func (l *Logger) writeFile(filename, message string) {
	l.fileMutex.Lock()
	defer l.fileMutex.Unlock()

	// Clean up file handles before opening a new file
	// to ensure the total number does not exceed the maximum limit
	if len(l.fileHandles) > l.maxFileHandles {
		l.cleanupFileHandles()
	}

	// Ensure the file handle is present and open
	file, ok := l.fileHandles[filename]
	if !ok || file == nil {
		var err error
		file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Failed to open log file: %v\n", err)
			return
		}
		l.fileHandles[filename] = file
	}

	// Update the access time for the file handle
	l.fileAccessTimes[filename] = time.Now()

	// Write the log message to the file
	if _, err := fmt.Fprintf(file, "%s\n", message); err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
		// Consider setting the file handle to nil on write failure
		l.fileHandles[filename] = nil
	}
}

// cleanupFileHandles closes and removes the least recently used file handles
// when the number of handles exceeds the maximum limit.
func (l *Logger) cleanupFileHandles() {
	for len(l.fileHandles) > l.maxFileHandles {
		oldestTime := time.Now()
		oldestFile := ""

		// Find the least recently used file handle
		for filename, accessTime := range l.fileAccessTimes {
			if accessTime.Before(oldestTime) {
				oldestTime = accessTime
				oldestFile = filename
			}
		}

		// Close and remove the oldest file handle
		if oldestFile != "" {
			if file, ok := l.fileHandles[oldestFile]; ok {
				if err := file.Close(); err != nil {
					fmt.Printf("Failed to close log file: %v\n", err)
				}
				delete(l.fileHandles, oldestFile)
				delete(l.fileAccessTimes, oldestFile)
			}
		}
	}
}

// cleanupUnusedFileHandles periodically closes file handles that have not been used for a certain duration.
func (l *Logger) cleanupUnusedFileHandles() {
	l.fileMutex.Lock()
	defer l.fileMutex.Unlock()

	threshold := time.Now().Add(-DefaultUnusedFileHandleThreshold)

	for filename, accessTime := range l.fileAccessTimes {
		if accessTime.Before(threshold) {
			if file, ok := l.fileHandles[filename]; ok {
				if err := file.Close(); err != nil {
					fmt.Printf("Failed to close log file: %v\n", err)
				}
				delete(l.fileHandles, filename)
				delete(l.fileAccessTimes, filename)
			}
		}
	}
}
