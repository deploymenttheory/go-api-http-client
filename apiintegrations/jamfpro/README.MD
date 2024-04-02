# Jamf Pro API Handler

The Jamf Pro API Handler is an integral component of the Go API HTTP Client, designed specifically for seamless integration with the Jamf Pro API. This handler facilitates the encoding and decoding of requests and responses, manages API-specific headers, and constructs endpoints for efficient API communication.

## Features

- **Endpoint Construction**: Dynamically constructs API resource and authentication endpoints based on the instance name and predefined URL patterns.
- **Content-Type Handling**: Determines the appropriate `Content-Type` header for requests, with specialized handling for both the Classic API (XML) and the JamfPro API (JSON).
- **Accept Header Management**: Generates a weighted `Accept` header to indicate the client's capability to process various MIME types, prioritizing XML for compatibility with the Classic API.
- **Standard Headers**: Provides a set of standard headers required for API requests, including `Accept`, `Content-Type`, and `Authorization`.
- **Request Marshaling**: Encodes request bodies into the appropriate format (XML or JSON) based on the target API endpoint, with support for multipart/form-data encoding for file uploads.
- **Response Handling**: Processes API responses, decoding them into the desired data structures and handling binary data responses where applicable.

The logic of this api handler is defined as follows:
Classic API:

For requests (GET, POST, PUT, DELETE):

- Encoding (Marshalling): Use XML format.
For responses (GET, POST, PUT):
- Decoding (Unmarshalling): Use XML format.
For responses (DELETE):
- Handle response codes as response body lacks anything useful.
Headers
- Sets accept headers based on weighting. XML out weighs JSON to ensure XML is returned
- Sets content header as application/xml with edge case exceptions based on need.

JamfPro API:

For requests (GET, POST, PUT, DELETE):

- Encoding (Marshalling): Use JSON format.
For responses (GET, POST, PUT):
- Decoding (Unmarshalling): Use JSON format.
For responses (DELETE):
- Handle response codes as response body lacks anything useful.
Headers
- Sets accept headers based on weighting. Jamf Pro API doesn't support XML, so MIME type is skipped and returns JSON
- Set content header as application/json with edge case exceptions based on need.


## Usage

To utilize the Jamf Pro API Handler within the Go API HTTP Client, instantiate the client with the Jamf Pro-specific configuration:

```go
package main

import (
    "log"

    "github.com/deploymenttheory/go-api-http-client/httpclient"
    "github.com/deploymenttheory/go-api-http-client/apihandlers/jamfpro"
)

func main() {
    // Configuration for the HTTP client specific to Jamf Pro
    config := httpclient.ClientConfig{
        Environment: httpclient.EnvironmentConfig{
            InstanceName: "your-instance-name",
            APIType:      "jamfpro", // Specify the API type as "jamfpro"
        },
        // Other configuration settings...
    }

    // Initialize the Jamf Pro API handler with the configuration
    jamfHandler, err := jamfpro.LoadAPIHandler(config.Environment.APIType, config.Logger)
    if err != nil {
        log.Fatalf("Failed to initialize Jamf Pro API handler: %v", err)
    }

    // Use the handler for API interactions
    // Example: Constructing an API resource endpoint
    resourceURL := jamfHandler.ConstructAPIResourceEndpoint(config.Environment.InstanceName, "/path/to/resource", config.Logger)
    log.Printf("Constructed Resource URL: %s", resourceURL)

    // Proceed with making API calls using the constructed URLs and the configured HTTP client...
}