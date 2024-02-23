# Go API HTTP Client

This Go module offers a sophisticated HTTP client designed for seamless API interactions, with a strong emphasis on concurrency management, robust error handling, extensive logging, and adaptive rate limiting. It's particularly suitable for applications requiring high-throughput API interactions with complex authentication and operational resilience.

## Features

- **Comprehensive Authentication Support**: Robust support for various authentication schemes, including OAuth and Bearer Token, with built-in token management and validation.
- **Advanced Concurrency Management**: An intelligent Concurrency Manager dynamically adjusts concurrent request limits to optimize throughput and adhere to API rate limits.
- **Structured Error Handling**: Clear and actionable error reporting facilitates troubleshooting and improves reliability.
- **Performance Monitoring**: Detailed performance metrics tracking provides insights into API interaction efficiency and optimization opportunities.
- **Configurable Logging**: Extensive logging capabilities with customizable levels and formats aid in debugging and operational monitoring.
- **Adaptive Rate Limiting**: Dynamic rate limiting automatically adjusts request rates in response to API server feedback.
- **Flexible Configuration**: Extensive customization of HTTP client behavior to meet specific API requirements, including custom timeouts, retry strategies, header management, and more.
- **Header Management**: Easy and efficient management of HTTP request headers, ensuring compliance with API requirements.
- **Enhanced Logging with Zap**: Utilizes Uber's zap library for structured, high-performance logging, offering levels from Debug to Fatal, including structured context and dynamic adjustment based on the environment.


## Getting Started

### Installation

To use this HTTP client in your project, add the package to your Go module dependencies:

```bash
go get github.com/yourusername/go-api-http-client
```

### Usage
Example usage with a configuration file:

```go
package main

import (
	"fmt"
	"log"

	"github.com/deploymenttheory/go-api-http-client/httpclient"
	"github.com/deploymenttheory/go-api-sdk-jamfpro/sdk/jamfpro"
)

func main() {
	configFilePath := "/path/to/clientconfig.json"
	loadedConfig, err := jamfpro.LoadClientConfig(configFilePath)
	if err != nil {
		log.Fatalf("Failed to load client OAuth configuration: %v", err)
	}

	config := httpclient.ClientConfig{
		Auth: httpclient.AuthConfig{
			ClientID:     loadedConfig.Auth.ClientID,
			ClientSecret: loadedConfig.Auth.ClientSecret,
		},
		Environment: httpclient.EnvironmentConfig{
			APIType:      loadedConfig.Environment.APIType,
			InstanceName: loadedConfig.Environment.InstanceName,
		},
		ClientOptions: httpclient.ClientOptions{
			LogLevel:          loadedConfig.ClientOptions.LogLevel,
			HideSensitiveData: loadedConfig.ClientOptions.HideSensitiveData,
			LogOutputFormat:   loadedConfig.ClientOptions.LogOutputFormat,
		},
	}

	client, err := jamfpro.BuildClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
}

```

Example configuration file (clientconfig.json):

```json
{
   "Auth": {
    "ClientID": "client-id",
    "ClientSecret": "client-secret",
    "Username": "username",
    "Password": "password"
  },
  "Environment": {
    "InstanceName": "yourinstance",
    "OverrideBaseDomain": "",
    "APIType": "" // "jamfpro" / "graph"
  },
  "ClientOptions": {
    "LogLevel": "LogLevelDebug", // "LogLevelDebug" / "LogLevelInfo" / "LogLevelWarn" / "LogLevelError" / "LogLevelFatal" / "LogLevelPanic"
    "LogOutputFormat": "console", // "console" / "json"
    "LogConsoleSeparator": " ", // " " / "\t" / "," / etc.
    "HideSensitiveData": true,  // true / false
  }
}
```

## Status

[![Super Linter](<https://github.com/segraef/Template/actions/workflows/linter.yml/badge.svg>)](<https://github.com/segraef/Template/actions/workflows/linter.yml>)

[![Sample Workflow](<https://github.com/segraef/Template/actions/workflows/workflow.yml/badge.svg>)](<https://github.com/segraef/Template/actions/workflows/workflow.yml>)


## Reporting Issues and Feedback

### Issues and Bugs

If you find any bugs, please file an issue in the [GitHub Issues][GitHubIssues] page. Please fill out the provided template with the appropriate information.

If you are taking the time to mention a problem, even a seemingly minor one, it is greatly appreciated, and a totally valid contribution to this project. **Thank you!**

## Feedback

Contributions are welcome to make this HTTP client even better! Feel free to fork the repository, make your improvements, and submit a pull request. For major changes or new features, please file an issue or feature request in the [GitHub Issues][GitHubIssues] page to discuss what you would like to change.

## Contribution

If you would like to become an active contributor to this repository or project, please follow the instructions provided in [`CONTRIBUTING.md`][Contributing].

## Learn More

* [GitHub Documentation][GitHubDocs]
* [Azure DevOps Documentation][AzureDevOpsDocs]
* [Microsoft Azure Documentation][MicrosoftAzureDocs]

<!-- References -->

<!-- Local -->
[ProjectSetup]: <https://docs.github.com/en/communities/setting-up-your-project-for-healthy-contributions>
[CreateFromTemplate]: <https://docs.github.com/en/github/creating-cloning-and-archiving-repositories/creating-a-repository-on-github/creating-a-repository-from-a-template>
[GitHubDocs]: <https://docs.github.com/>
[AzureDevOpsDocs]: <https://docs.microsoft.com/en-us/azure/devops/?view=azure-devops>
[GitHubIssues]: <https://github.com/segraef/Template/issues>
[Contributing]: CONTRIBUTING.md

<!-- External -->
[Az]: <https://img.shields.io/powershellgallery/v/Az.svg?style=flat-square&label=Az>
[AzGallery]: <https://www.powershellgallery.com/packages/Az/>
[PowerShellCore]: <https://github.com/PowerShell/PowerShell/releases/latest>

<!-- Docs -->
[MicrosoftAzureDocs]: <https://docs.microsoft.com/en-us/azure/>
[PowerShellDocs]: <https://docs.microsoft.com/en-us/powershell/>
