# Go API HTTP Client

This Go module offers a sophisticated HTTP client designed for seamless API interactions, with a strong emphasis on concurrency management, robust error handling, extensive logging, and adaptive rate limiting. It's particularly suitable for applications requiring high-throughput API interactions with complex authentication and operational resilience.

This client is designed to be used with targetted SDK's and terraform providers only. As such the http client cannot be used without a supporting SDK and associated api integration plugin [go-api-http-client-integrations](https://github.com/deploymenttheory/go-api-http-client-integrations).

The plugin is required to provide the necessary API-specific handlers and configuration to the HTTP client. The plugin approach is designed to provide a consistent interface for interacting with various APIs, including Microsoft Graph, Jamf Pro, and others. It is easily extensible to support additional APIs and highly configurable to meet specific API requirements. It achieves this through using a modular design, with a core HTTP client and API-specific handlers that encapsulate the unique requirements of each API supported. Conseqently the client provides core common HTTP client functionality, such as rate limiting, logging, concurrency and responses while the plugin provides the API-specific logic, such as encoding and decoding requests, managing authentication endpoints, and handling API-specific logic.

## HTTP Client Features

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

## Getting Started

This SDK requires an existing Go environment to build and run. If you do not have Go installed, you can download and install it from the [official Go website](https://golang.org/doc/install).



## SDK Support

This http client is used in conjuction with the following SDKs:

- [go-api-sdk-m365](https://github.com/deploymenttheory/go-api-sdk-m365)

- [go-api-sdk-jamfpro](https://github.com/deploymenttheory/go-api-sdk-jamfpro)

## Reporting Issues and Feedback

### Issues and Bugs

If you find any bugs, please file an issue in the [GitHub Issues][GitHubIssues] page. Please fill out the provided template with the appropriate information.

If you are taking the time to mention a problem, even a seemingly minor one, it is greatly appreciated, and a totally valid contribution to this project. **Thank you!**

## Feedback

Contributions are welcome to make this HTTP client even better! Feel free to fork the repository, make your improvements, and submit a pull request. For major changes or new features, please file an issue or feature request in the [GitHub Issues][GitHubIssues] page to discuss what you would like to change.

## Contribution

If you would like to become an active contributor to this repository or project, please follow the instructions provided in [`CONTRIBUTING.md`][Contributing].

## Learn More

<!-- References -->

<!-- Local -->
[ProjectSetup]: <https://docs.github.com/en/communities/setting-up-your-project-for-healthy-contributions>
[CreateFromTemplate]: <https://docs.github.com/en/github/creating-cloning-and-archiving-repositories/creating-a-repository-on-github/creating-a-repository-from-a-template>
[GitHubDocs]: <https://docs.github.com/>
[AzureDevOpsDocs]: <https://docs.microsoft.com/en-us/azure/devops/?view=azure-devops>
[GitHubIssues]: <https://github.com/segraef/Template/issues>
[Contributing]: CONTRIBUTING.md