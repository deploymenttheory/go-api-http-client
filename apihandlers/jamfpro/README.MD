------------------------------Summary----------------------------------------
This module is a api handler module for the go-api-http-client to accommodate specifics of
jamf's api(s). It handles the encoding (marshalling) and decoding (unmarshalling)
of data. It also sets the correct content headers for the various http methods.

This module integrates with the go-api-http-client logger for wrapped error handling
for human readable return codes. It also supports the go-api-http-client tiered logging
functionality for logging support.

The logic of this module is defined as follows:
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
