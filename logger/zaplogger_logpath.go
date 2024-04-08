// logger/zaplogger_logpath.go

package logger

import (
	"os"
	"path/filepath"
	"time"
)

// EnsureLogFilePath checks the provided path and prepares it for use with the logger.
// If the path is a directory, it appends a timestamp-based filename.
// If the path includes a filename, it checks for the existence of the file.
// If no path is provided, it defaults to creating a log file in the current directory with a timestamp-based name.
func EnsureLogFilePath(logPath string) (string, error) {
	if logPath == "" {
		// Default to the current directory with a timestamp-based filename if no path is provided
		logPath = filepath.Join(".", "log_"+time.Now().Format("20060102_150405")+".log")
	} else {
		info, err := os.Stat(logPath)

		if os.IsNotExist(err) || (err == nil && info.IsDir()) {
			// If the path doesn't exist or is a directory, append a timestamp-based filename
			logPath = filepath.Join(logPath, "log_"+time.Now().Format("20060102_150405")+".log")
		} else if err != nil {
			// If there's an error other than "not exists", return it
			return "", err
		}
		// If the path exists and is not a directory, it's assumed to be a filename and will be used as is
	}

	// Ensure the directory exists
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	return logPath, nil
}
