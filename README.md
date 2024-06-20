# Go API HTTP Client

This Go module offers a sophisticated HTTP client designed for seamless API interactions, with a strong emphasis on concurrency management, robust error handling, extensive logging, and adaptive rate limiting. It's particularly suitable for applications requiring high-throughput API interactions with complex authentication and operational resilience.

This client leverages API-specific SDKs to provide a comprehensive and consistent interface for interacting with various APIs, including Microsoft Graph, Jamf Pro, and others. It is designed to be easily extensible to support additional APIs and to be highly configurable to meet specific API requirements. It achieves this through using a modular design, with a core HTTP client and API-specific handlers that encapsulate the unique requirements of each API supported.

This HTTP client is intended to be used with targetted SDK's and terraform providers only. As such the http client cannot be used without a supporting SDK.

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
- **API Handler Interface**: Provides a flexible and extensible way to interact with different APIs, including encoding and decoding requests and responses, managing authentication endpoints, and handling API-specific logic.
- **Configuration via JSON or Environment Variables**: The Go API HTTP Client supports configuration via JSON files or environment variables, providing flexibility in defining authentication credentials, API endpoints, logging settings, and other parameters.

TBC

## Getting Started
'
{
	"Integration": {
		"FQDN": "https://api.example.com",
		"AuthMethodDescriptor": "Bearer"
		// other necessary fields for integration
	},
	"HideSensitiveData": true,
	"CustomCookies": [
		{
			"Name": "sessionid",
			"Value": "abc123",
			"Path": "/",
			"Domain": "example.com",
			"Expires": "2025-01-02T15:04:05Z",
			"MaxAge": 86400,
			"Secure": true,
			"HttpOnly": true,
			"SameSite": 1
		}
	],
	"MaxRetryAttempts": 5,
	"MaxConcurrentRequests": 10,
	"EnableDynamicRateLimiting": true,
	"CustomTimeout": "10s",
	"TokenRefreshBufferPeriod": "2m",
	"TotalRetryDuration": "10m",
	"FollowRedirects": true,
	"MaxRedirects": 3,
	"ConcurrencyManagementEnabled": true
}
'


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