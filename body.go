package httpman

import (
	"bytes"
	"io"
	"strings"

	goquery "github.com/google/go-querystring/query"
	jsoniter "github.com/json-iterator/go"
)

const (
	contentType     = "Content-Type"
	jsonContentType = "application/json"
	formContentType = "application/x-www-form-urlencoded"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// BodyProvider provides Body content for http.Request attachment.
type BodyProvider interface {
	// ContentType returns the Content-Type of the body.
	ContentType() string
	// Body returns the io.Reader body.
	Body() (io.Reader, error)
}

// bodyProvider provides the wrapped body value as a Body for requests.
type bodyProvider struct {
	body io.Reader
}

func (p *bodyProvider) ContentType() string {
	return ""
}

func (p *bodyProvider) Body() (io.Reader, error) {
	return p.body, nil
}

// jsonBodyProvider encodes a JSON tagged struct value as a Body for requests.
type jsonBodyProvider struct {
	payload interface{}
}

func (p *jsonBodyProvider) ContentType() string {
	return jsonContentType
}

func (p *jsonBodyProvider) Body() (io.Reader, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(p.payload)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// formBodyProvider encodes a url tagged struct value as Body for requests.
type formBodyProvider struct {
	payload interface{}
}

func (p *formBodyProvider) ContentType() string {
	return formContentType
}

func (p *formBodyProvider) Body() (io.Reader, error) {
	values, err := goquery.Values(p.payload)
	if err != nil {
		return nil, err
	}
	return strings.NewReader(values.Encode()), nil
}
