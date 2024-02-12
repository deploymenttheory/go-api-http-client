package jamfpro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

/*
// ExtractErrorMessageFromHTML attempts to parse an HTML error page and extract a human-readable error message.
func ExtractErrorMessageFromHTML(htmlContent string) string {
	r := bytes.NewReader([]byte(htmlContent))
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "Unable to parse HTML content"
	}

	var messages []string
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		messages = append(messages, s.Text())
	})

	return strings.Join(messages, " | ")
}
*/

// ExtractErrorMessageFromHTML attempts to parse an HTML error page and extract human-readable error messages as key-value pairs.
func ExtractErrorMessageFromHTML(htmlContent string) []map[string]string {
	r := bytes.NewReader([]byte(htmlContent))
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return []map[string]string{{"Error": "Unable to parse HTML content"}}
	}

	var messages []map[string]string
	errorCount := 1
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		key := s.Find("strong").First().Text()
		value := strings.TrimSpace(s.Clone().Children().Remove().End().Text())

		if key == "" {
			key = fmt.Sprintf("Error %d", errorCount)
			errorCount++
		}

		if value == "" {
			value = s.Text()
		}

		messages = append(messages, map[string]string{key: value})
	})

	return messages
}

// ParseJSONErrorResponse parses the JSON error message from the response body.
func ParseJSONErrorResponse(body []byte) (string, error) {
	var errorResponse struct {
		HTTPStatus int `json:"httpStatus"`
		Errors     []struct {
			Code        string `json:"code"`
			Description string `json:"description"`
			ID          string `json:"id"`
			Field       string `json:"field"`
		} `json:"errors"`
	}

	err := json.Unmarshal(body, &errorResponse)
	if err != nil {
		return "", err
	}

	if len(errorResponse.Errors) > 0 {
		return errorResponse.Errors[0].Description, nil
	}

	return "No error description available", nil
}
