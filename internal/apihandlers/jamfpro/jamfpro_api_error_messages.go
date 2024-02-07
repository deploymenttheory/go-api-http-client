package jamfpro

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

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
