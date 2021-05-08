package aur

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type PayloadError struct {
	StatusCode int
	ErrorField string
}

func (r *PayloadError) Error() string {
	return fmt.Sprintf("status %d: err %s", r.StatusCode, r.ErrorField)
}

const _defaultURL = "https://aur.archlinux.org/rpc.php?"

// The interface specification for the client above.
type ClientInterface interface {
	// Search queries the AUR DB with an optional By filter.
	// Use By.None for default query param (name-desc)
	Search(ctx context.Context, query string, by By, reqEditors ...RequestEditorFn) ([]Pkg, error)

	// Info gives detailed information on existing package.
	Info(ctx context.Context, pkgs []string, reqEditors ...RequestEditorFn) ([]Pkg, error)
}

// Client for AUR searching and querying.
type Client struct {
	baseURL string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	HTTPClient HTTPRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction.
type ClientOption func(*Client) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HTTPRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// RequestEditorFn  is the function signature for the RequestEditor callback function.
type RequestEditorFn func(ctx context.Context, req *http.Request) error

func NewClient(opts ...ClientOption) (*Client, error) {
	client := Client{
		baseURL:        "",
		HTTPClient:     nil,
		RequestEditors: []RequestEditorFn{},
	}

	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}

	// create httpClient, if not already present
	if client.HTTPClient == nil {
		client.HTTPClient = http.DefaultClient
	}

	// set default baseURL if not present or valid
	if client.baseURL == "" {
		client.baseURL = _defaultURL
	}

	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.baseURL, "/") {
		client.baseURL += "/"
	}

	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HTTPRequestDoer) ClientOption {
	return func(c *Client) error {
		c.HTTPClient = doer

		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)

		return nil
	}
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}

	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}

	return nil
}

func newAURRPCRequest(baseURL string, values url.Values) (*http.Request, error) {
	values.Set("v", "5")

	req, err := http.NewRequest("GET", baseURL+values.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

func parseRPCResponse(resp *http.Response) ([]Pkg, error) {
	defer resp.Body.Close()

	if err := getErrorByStatusCode(resp.StatusCode); err != nil {
		return nil, err
	}

	result := new(response)

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, fmt.Errorf("response decoding failed: %w", err)
	}

	if len(result.Error) > 0 {
		return nil, &PayloadError{
			StatusCode: resp.StatusCode,
			ErrorField: result.Error,
		}
	}

	return result.Results, nil
}

// Search queries the AUR DB with an optional By field.
// Use By.None for default query param (name-desc)
func (c *Client) Search(ctx context.Context, query string, by By, reqEditors ...RequestEditorFn) ([]Pkg, error) {
	v := url.Values{}
	v.Set("type", "search")
	v.Set("arg", query)

	if by != None {
		v.Set("by", by.String())
	}

	return c.get(ctx, v, reqEditors)
}

// Info shows info for one or multiple packages.
func (c *Client) Info(ctx context.Context, pkgs []string, reqEditors ...RequestEditorFn) ([]Pkg, error) {
	v := url.Values{}
	v.Set("type", "info")

	for _, arg := range pkgs {
		v.Add("arg[]", arg)
	}

	return c.get(ctx, v, reqEditors)
}

func (c *Client) get(ctx context.Context, values url.Values, reqEditors []RequestEditorFn) ([]Pkg, error) {
	req, err := newAURRPCRequest(c.baseURL, values)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return parseRPCResponse(resp)
}
