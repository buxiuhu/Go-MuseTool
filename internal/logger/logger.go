package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	LogDir      = "logs"
	MaxLogDays  = 30
	LogFileName = "app_%s.log" // app_2023-10-27.log
)

var (
	infoLogger   *log.Logger
	errorLogger  *log.Logger
	debugLogger  *log.Logger
	logFile      *os.File
	logMutex     sync.Mutex
	debugEnabled bool
)

// Setup initializes the logging system.
// It creates the log directory, sets up the log file for the current day,
// and cleans up old logs.
func Setup(enableDebug bool) error {
	logMutex.Lock()
	defer logMutex.Unlock()

	debugEnabled = enableDebug

	// 1. Create logs directory if not exists
	if err := os.MkdirAll(LogDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// 2. Open log file for today
	today := time.Now().Format("2006-01-02")
	logFilePath := filepath.Join(LogDir, fmt.Sprintf(LogFileName, today))

	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// 3. Set log output to both file and stdout
	mw := io.MultiWriter(os.Stdout, logFile)

	// Initialize different loggers
	// Format: 【LEVEL】YYYY/MM/DD HH:MM:SS message
	infoLogger = log.New(mw, "【INFO】", log.Ldate|log.Ltime)
	errorLogger = log.New(mw, "【ERROR】", log.Ldate|log.Ltime|log.Lshortfile)

	if debugEnabled {
		debugLogger = log.New(mw, "【DEBUG】", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		debugLogger = log.New(io.Discard, "【DEBUG】", log.Ldate|log.Ltime|log.Lshortfile)
	}

	Info("Logging initialized. Writing to %s (Debug: %v)", logFilePath, debugEnabled)

	// 4. Clean up old logs
	go cleanOldLogs()

	return nil
}

// Close closes the log file
func Close() {
	logMutex.Lock()
	defer logMutex.Unlock()

	Info("Application shutting down")

	if logFile != nil {
		logFile.Close()
	}
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	if infoLogger != nil {
		infoLogger.Printf(format, v...)
	} else {
		log.Printf("【INFO】"+format, v...)
	}
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	if errorLogger != nil {
		errorLogger.Printf(format, v...)
	} else {
		log.Printf("【ERROR】"+format, v...)
	}
}

// Debug logs a debug message (only if debug is enabled)
func Debug(format string, v ...interface{}) {
	if debugLogger != nil && debugEnabled {
		debugLogger.Printf(format, v...)
	}
}

// SetDebugEnabled enables or disables debug logging
func SetDebugEnabled(enabled bool) {
	logMutex.Lock()
	defer logMutex.Unlock()

	debugEnabled = enabled

	if logFile != nil {
		mw := io.MultiWriter(os.Stdout, logFile)
		if enabled {
			debugLogger = log.New(mw, "【DEBUG】", log.Ldate|log.Ltime|log.Lshortfile)
			Info("Debug logging enabled")
		} else {
			debugLogger = log.New(io.Discard, "【DEBUG】", log.Ldate|log.Ltime|log.Lshortfile)
			Info("Debug logging disabled")
		}
	}
}

// IsDebugEnabled returns whether debug logging is enabled
func IsDebugEnabled() bool {
	logMutex.Lock()
	defer logMutex.Unlock()
	return debugEnabled
}

func cleanOldLogs() {
	files, err := os.ReadDir(LogDir)
	if err != nil {
		Error("Failed to read log directory for cleanup: %v", err)
		return
	}

	var logFiles []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".log" {
			logFiles = append(logFiles, filepath.Join(LogDir, file.Name()))
		}
	}

	// Sort files by name (which includes date)
	sort.Strings(logFiles)

	// Delete old files if we have more than MaxLogDays
	if len(logFiles) > MaxLogDays {
		filesToDelete := logFiles[:len(logFiles)-MaxLogDays]
		for _, f := range filesToDelete {
			if err := os.Remove(f); err != nil {
				Error("Failed to delete old log file %s: %v", f, err)
			} else {
				Info("Deleted old log file: %s", f)
			}
		}
	}
}
