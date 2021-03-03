package httpman

import (
	goquery "github.com/google/go-querystring/query"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Request
type Request struct {
	// request method
	method string
	// request absolute URL
	absoluteURL string
	// request header collection
	header http.Header
	// http client instance
	httpmanInstance *Httpman
	// body provider
	bodyProvider BodyProvider
	// url query structs
	queryStructs []interface{}
	// url query map
	queryMap map[string]string
	// response decoder
	responseDecoder ResponseDecoder
}

// initiates a new request with defaults
func (h *Httpman) NewRequest() *Request {
	return &Request{
		httpmanInstance: h,
		queryStructs:    make([]interface{}, 0),
		queryMap:        make(map[string]string),
		header:          make(http.Header),
		responseDecoder: jsonDecoder{},
		method:          http.MethodGet,
		absoluteURL:     h.baseURL,
	}
}

// Method

// Head sets method to HEAD and sets the given pathURL.
func (r *Request) Head(pathURL string) *Request {
	r.method = http.MethodHead
	return r.Path(pathURL)
}

// Get sets method to GET and sets the given pathURL.
func (r *Request) Get(pathURL string) *Request {
	r.method = http.MethodGet
	return r.Path(pathURL)
}

// Post sets method to POST and sets the given pathURL.
func (r *Request) Post(pathURL string) *Request {
	r.method = http.MethodPost
	return r.Path(pathURL)
}

// Put sets method to PUT and sets the given pathURL.
func (r *Request) Put(pathURL string) *Request {
	r.method = http.MethodPut
	return r.Path(pathURL)
}

// Patch sets method to PATCH and sets the given pathURL.
func (r *Request) Patch(pathURL string) *Request {
	r.method = http.MethodPatch
	return r.Path(pathURL)
}

// Delete sets method to DELETE and sets the given pathURL.
func (r *Request) Delete(pathURL string) *Request {
	r.method = http.MethodDelete
	return r.Path(pathURL)
}

// Options sets method to OPTIONS and sets the given pathURL.
func (r *Request) Options(pathURL string) *Request {
	r.method = http.MethodOptions
	return r.Path(pathURL)
}

// Trace sets method to TRACE and sets the given pathURL.
func (r *Request) Trace(pathURL string) *Request {
	r.method = http.MethodTrace
	return r.Path(pathURL)
}

// Connect sets method to CONNECT and sets the given pathURL.
func (r *Request) Connect(pathURL string) *Request {
	r.method = http.MethodConnect
	return r.Path(pathURL)
}

// Path extends the rawURL with the given path by resolving the reference to
// an absolute URL. If parsing errors occur, the baseURL is left unmodified.
func (r *Request) Path(path string) *Request {
	baseURL, baseErr := url.Parse(r.httpmanInstance.baseURL)
	pathURL, pathErr := url.Parse(path)
	if baseErr == nil && pathErr == nil {
		r.absoluteURL = baseURL.ResolveReference(pathURL).String()
		return r
	}
	return r
}

// QueryStruct appends the queryStruct to the queryStructs. The value
// pointed to by each queryStruct will be encoded as url query parameters on
// The queryStruct argument should be a pointer to a url tagged struct.
func (r *Request) AddQueryStruct(queryStruct interface{}) *Request {
	if queryStruct != nil {
		r.queryStructs = append(r.queryStructs, queryStruct)
	}
	return r
}

// AddQueryParam appends query param to the queryStruct. The value
// pointed to by each queryStruct will be encoded as url query parameters on
func (r *Request) AddQueryParam(key, value string) *Request {
	if len(key) > 0 && len(value) > 0 {
		r.queryMap[key] = value
	}
	return r
}

// Body

// Body sets the body. The body value will be set as the Body on new
// If the provided body is also an io.Closer, the request Body will be closed
// by http.Client methods.
func (r *Request) Body(body io.Reader) *Request {
	if body == nil {
		return r
	}
	return r.BodyProvider(&bodyProvider{body: body})
}

// BodyProvider sets the body provider.
func (r *Request) BodyProvider(body BodyProvider) *Request {
	if body == nil {
		return r
	}
	r.bodyProvider = body

	ct := body.ContentType()
	if ct != "" {
		r.SetHeader(contentType, ct)
	}

	return r
}

// SetHeader sets the key, value pair in Headers, replacing existing values
// associated with key. Header keys are canonicalize.
func (r *Request) SetHeader(key, value string) *Request {
	r.header.Set(key, value)
	return r
}

// BodyJSON sets the bodyJSON. The value pointed to by the bodyJSON
// will be JSON encoded as the Body on new requests.
// The bodyJSON argument should be a pointer to a JSON tagged struct.
func (r *Request) BodyJSON(bodyJSON interface{}) *Request {
	if bodyJSON == nil {
		return r
	}
	return r.BodyProvider(&jsonBodyProvider{payload: bodyJSON})
}

// BodyForm sets the bodyForm. The value pointed to by the bodyForm
// will be url encoded as the Body on new requests.
// The bodyForm argument should be a pointer to a url tagged struct.
func (r *Request) BodyForm(bodyForm interface{}) *Request {
	if bodyForm == nil {
		return r
	}
	return r.BodyProvider(&formBodyProvider{payload: bodyForm})
}

func (r *Request) Send() (*http.Request, error) {
	reqURL, err := url.Parse(r.absoluteURL)
	if err != nil {
		return nil, err
	}

	if len(r.httpmanInstance.queryStructs) != 0 {
		r.queryStructs = append(r.queryStructs, r.httpmanInstance.queryStructs...)
	}

	err = addQueryStructs(reqURL, r.queryStructs)
	if err != nil {
		return nil, err
	}

	var body io.Reader
	if r.bodyProvider != nil {
		body, err = r.bodyProvider.Body()
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(r.method, reqURL.String(), body)
	if err != nil {
		return nil, err
	}
	addQueryMap(req, r.httpmanInstance.queryMap, r.queryMap)
	addHeaders(req, r.httpmanInstance.header, r.header)
	return req, err
}

// ReceiveSuccess creates a new HTTP request and returns the response. Success
// responses (2XX) are JSON decoded into the value pointed to by successV.
// Any error creating the request, sending it, or decoding a 2XX response
// is returned.
func (r *Request) DecodeSuccess(successV interface{}) (*http.Response, error) {
	return r.Decode(successV, nil)
}

// Receive creates a new HTTP request and returns the response. Success
// responses (2XX) are JSON decoded into the value pointed to by successV and
// other responses are JSON decoded into the value pointed to by failureV.
// If the status code of response is 204(no content) or the Content-Length is 0,
// decoding is skipped. Any error creating the request, sending it, or decoding
// the response is returned.
// Receive is shorthand for calling Request and Do.
func (r *Request) Decode(successV, failureV interface{}) (*http.Response, error) {
	req, err := r.Send()
	if err != nil {
		return nil, err
	}
	return r.Do(req, successV, failureV)
}

// Do sends an HTTP request and returns the response. Success responses (2XX)
// are JSON decoded into the value pointed to by successV and other responses
// are JSON decoded into the value pointed to by failureV.
// If the status code of response is 204(no content) or the Content-Length is 0,
// decoding is skipped. Any error sending the request or decoding the response
// is returned.
func (r *Request) Do(req *http.Request, successV, failureV interface{}) (*http.Response, error) {
	resp, err := r.httpmanInstance.httpClient.Do(req)
	if err != nil {
		return resp, err
	}
	// when err is nil, resp contains a non-nil resp.Body which must be closed
	defer resp.Body.Close()

	// The default HTTP client's Transport may not
	// reuse HTTP/1.x "keep-alive" TCP connections if the Body is
	// not read to completion and closed.
	// See: https://golang.org/pkg/net/http/#Response
	defer io.Copy(ioutil.Discard, resp.Body)

	// Don't try to decode on 204s or Content-Length is 0
	if resp.StatusCode == http.StatusNoContent || resp.ContentLength == 0 {
		return resp, nil
	}

	// Decode from json
	if successV != nil || failureV != nil {
		err = decodeResponse(resp, r.responseDecoder, successV, failureV)
	}
	return resp, err
}

// decodeResponse decodes response Body into the value pointed to by successV
// if the response is a success (2XX) or into the value pointed to by failureV
// otherwise. If the successV or failureV argument to decode into is nil,
// decoding is skipped.
// Caller is responsible for closing the resp.Body.
func decodeResponse(resp *http.Response, decoder ResponseDecoder, successV, failureV interface{}) error {
	if code := resp.StatusCode; 200 <= code && code <= 299 {
		if successV != nil {
			return decoder.Decode(resp, successV)
		}
	} else {
		if failureV != nil {
			return decoder.Decode(resp, failureV)
		}
	}
	return nil
}

// addQueryStructs parses url tagged query structs using go-querystring to
// encode them to url.Values and format them onto the url.RawQuery. Any
// query parsing or encoding errors are returned.
func addQueryStructs(reqURL *url.URL, queryStructs []interface{}) error {
	urlValues, err := url.ParseQuery(reqURL.RawQuery)
	if err != nil {
		return err
	}
	// encodes query structs into a url.Values map and merges maps
	for _, queryStruct := range queryStructs {
		queryValues, err := goquery.Values(queryStruct)
		if err != nil {
			return err
		}
		for key, values := range queryValues {
			for _, value := range values {
				urlValues.Add(key, value)
			}
		}
	}
	// url.Values format to a sorted "url encoded" string, e.g. "key=val&foo=bar"
	reqURL.RawQuery = urlValues.Encode()
	return nil
}

// addHeaders adds the key, value pairs from the given http.Header to the
// request. Values for existing keys are appended to the keys values.
func addHeaders(req *http.Request, defaultHeaders, headers http.Header) {
	for key, values := range defaultHeaders {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
}

func addQueryMap(req *http.Request, defaultQueryMap map[string]string, queryMap map[string]string) {
	if len(defaultQueryMap) != 0 || len(queryMap) != 0 {
		q := req.URL.Query()
		for key, value := range defaultQueryMap {
			q.Add(key, value)
		}
		for key, value := range queryMap {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}
}
