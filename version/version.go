// version.go
package version

// AppName holds the name of the application
var AppName = "go-api-http-client"

// Version holds the current version of the application
var Version = "0.0.65"

// GetAppName returns the name of the application
func GetAppName() string {
	return AppName
}

// GetVersion returns the current version of the application
func GetVersion() string {
	return Version
}
