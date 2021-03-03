package httpman

import (
	"encoding/base64"
	"net/http"
)

// Executor executes http requests.  It is implemented by *http.Client.  You can
// wrap *http.Client with layers of Doers to form a stack of client-side
// middleware.
type Executor interface {
	Do(req *http.Request) (*http.Response, error)
}

// Httpman is an HTTP Request builder and sender.
type Httpman struct {
	// http Client for executing requests
	httpClient Executor
	// raw url string for requests
	baseURL string
	// stores key-values pairs to add to request's Headers
	header http.Header
	// url query structs
	queryStructs []interface{}
	// url query map
	queryMap map[string]string
}

// New returns a new instance with an http DefaultClient.
func New(baseURL string) *Httpman {
	return &Httpman{
		httpClient:   http.DefaultClient,
		header:       make(http.Header),
		baseURL:      baseURL,
		queryStructs: make([]interface{}, 0),
		queryMap:     make(map[string]string),
	}
}

// Http Client

// Client sets the http Client used to do requests. If a nil client is given,
// the http.DefaultClient will be used.
func (h *Httpman) Client(httpClient *http.Client) *Httpman {
	if httpClient == nil {
		return h.Doer(http.DefaultClient)
	}
	return h.Doer(httpClient)
}

// Doer sets the custom Doer implementation used to do requests.
// If a nil client is given, the http.DefaultClient will be used.
func (h *Httpman) Doer(doer Executor) *Httpman {
	if doer == nil {
		h.httpClient = http.DefaultClient
	} else {
		h.httpClient = doer
	}
	return h
}

// Header

// AddHeader adds the key, value pair in Headers, appending values for existing keys
// to the key's values. Header keys are canonicalize.
func (h *Httpman) AddHeader(key, value string) *Httpman {
	h.header.Add(key, value)
	return h
}

// SetHeader sets the key, value pair in Headers, replacing existing values
// associated with key. Header keys are canonicalize.
func (h *Httpman) SetHeader(key, value string) *Httpman {
	h.header.Set(key, value)
	return h
}

// SetBasicAuth sets the Authorization header to use HTTP Basic Authentication
// with the provided username and password. With HTTP Basic Authentication
// the provided username and password are not encrypted.
func (h *Httpman) SetBasicAuth(username, password string) *Httpman {
	return h.SetHeader("Authorization", "Basic "+basicAuth(username, password))
}

// QueryStruct appends the queryStruct to the queryStructs. The value
// pointed to by each queryStruct will be encoded as url query parameters on
// The queryStruct argument should be a pointer to a url tagged struct.
func (h *Httpman) AddQueryStruct(queryStruct interface{}) *Httpman {
	if queryStruct != nil {
		h.queryStructs = append(h.queryStructs, queryStruct)
	}
	return h
}

// AddQueryParam appends query param to the queryStruct. The value
// pointed to by each queryStruct will be encoded as url query parameters on
func (h *Httpman) AddQueryParam(key, value string) *Httpman {
	if len(key) > 0 && len(value) > 0 {
		h.queryMap[key] = value
	}
	return h
}

// basicAuth returns the base64 encoded username:password for basic auth copied
// from net/http.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
