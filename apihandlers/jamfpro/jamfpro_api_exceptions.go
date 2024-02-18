package jamfpro

import (
	_ "embed"

	"encoding/json"
	"log"
)

// ConfigMap is a map that associates endpoint URL patterns with their corresponding configurations.
// The map's keys are strings that identify the endpoint, and the values are EndpointConfig structs
// that hold the configuration for that endpoint.
type ConfigMap map[string]EndpointConfig

// Variables
var configMap ConfigMap

// Embedded Resources
//
//go:embed jamfpro_api_exceptions_configuration.json
var jamfpro_api_exceptions_configuration []byte

// init is invoked automatically on package initialization and is responsible for
// setting up the default state of the package by loading the api exceptions configuration.
func init() {
	// Load the default configuration from an embedded resource.
	err := loadAPIExceptionsConfiguration()
	if err != nil {
		log.Fatalf("Error loading Jamf Pro API exceptions configuration: %s", err)
	}
}

// loadAPIExceptionsConfiguration reads and unmarshals the jamfpro_api_exceptions_configuration JSON data from an embedded file
// into the configMap variable, which holds the exceptions configuration for endpoint-specific headers.
func loadAPIExceptionsConfiguration() error {
	// Unmarshal the embedded default configuration into the global configMap.
	return json.Unmarshal(jamfpro_api_exceptions_configuration, &configMap)
}
