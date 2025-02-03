package monitor

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Monitor struct {
	file *os.File
	*log.Logger
	mu sync.Mutex
}

// New creates a new Monitor instance that writes logs to the specified file.
// If the file does not exist, it will be created. If there is an error opening
// the file, the function returns nil. The Logger is initialized with standard
// log flags. The returned Monitor must be closed using the Close method when
// it is no longer needed to ensure the file is properly closed.

func New(filename string) *Monitor {

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	if err != nil {
		return nil
	}

	return &Monitor{
		file:   file,
		Logger: log.New(file, "", log.LstdFlags),
	}
}

// Log writes a message to the log with the given level. The level can be any
// string, but common levels are "INFO", "ERROR", "DEBUG", and "WARNING". The
// message is formatted with a timestamp and the log level, and then written to
// the underlying logger.
func (m *Monitor) Log(level, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	loggerMessage := fmt.Sprintf("[%s] [%s] %s ", timestamp, level, message)

	m.Logger.Println(loggerMessage)
}

// Close closes the file associated with the Monitor.
// It must be called when the Monitor is no longer needed to ensure
// that the file is properly closed and resources are released.

func (m *Monitor) Close() {
	m.file.Close()
}

// Info logs a message with the level "INFO".
func (m *Monitor) Info(message string) {
	m.Log("INFO", message)
}

// Error logs a message with the level "ERROR".

// Error logs a message with the level "ERROR".
// This method formats the message with a timestamp and the log level,
// then writes it to the underlying logger.

func (m *Monitor) Error(message string) {
	m.Log("ERROR", message)
}

// Debug logs a message with the level "DEBUG".
// This method formats the message with a timestamp and the log level,
// then writes it to the underlying logger.

func (m *Monitor) Debug(message string) {
	m.Log("DEBUG", message)
}

// Warning logs a message with the level "WARNING".
// This method formats the message with a timestamp and the log level,
// then writes it to the underlying logger.
func (m *Monitor) Warning(message string) {
	m.Log("WARNING", message)
}
