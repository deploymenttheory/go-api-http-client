package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type cookieWithJson struct{}

type configProcessedVals struct {
	CustomCookies                   []cookieWithJson `json:"custom_cookies>cookie"`
	CustomTimeoutSeconds            int              `json:"custom_timeout_seconds"`
	TokenRefreshBufferPeriodSeconds int              `json:"token_refresh_buffer_period_seconds"`
	TotalRetryDurationSeconds       int              `json:"total_retry_duration_seconds"`
}

func BuildConfigFromJsonFile(filepath string) (ClientConfig, error) {
	var clientConfig ClientConfig
	var processorContainer *configProcessedVals

	// Load file to base struct
	loadConfigFromJSONFile(filepath, clientConfig)

	// Load file to processor struct
	loadConfigFromJSONFile(filepath, processorContainer)

	return clientConfig, nil
}

// loadCombinedConfig loads the combined configuration from a JSON file
func loadConfigFromJSONFile(configFilePath string, home any) error {
	file, err := os.Open(configFilePath)
	if err != nil {
		return fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("could not read file: %v", err)
	}

	err = json.Unmarshal(byteValue, &home)
	if err != nil {
		return fmt.Errorf("could not unmarshal JSON: %v", err)
	}

	return nil
}
