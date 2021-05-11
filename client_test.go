package aur

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const errorPayload = `{"version":5,"type":"error","resultcount":0,
"results":[],"error":"Incorrect by field specified."}`

const noMatchPayload = `{"version":5,"type":"error","resultcount":0,
"results":[],"error":""}`

const validPayload = `{
    "version":5,
    "type":"multiinfo",
    "resultcount":1,
    "results":[{
        "ID":229417,
        "Name":"cower",
        "PackageBaseID":44921,
        "PackageBase":"cower",
        "Version":"14-2",
        "Description":"A simple AUR agent with a pretentious name",
        "URL":"http:\/\/github.com\/falconindy\/cower",
        "NumVotes":590,
        "Popularity":24.595536,
        "OutOfDate":null,
        "Maintainer":"falconindy",
        "FirstSubmitted":1293676237,
        "LastModified":1441804093,
        "URLPath":"\/cgit\/aur.git\/snapshot\/cower.tar.gz",
        "Depends":[
            "curl",
            "openssl",
            "pacman",
            "yajl"
        ],
        "MakeDepends":[
            "perl"
        ],
        "License":[
            "MIT"
        ],
        "Keywords":[]
    }]
 }
`

var validPayloadItems = []Pkg{{ID: 229417, Name: "cower", PackageBaseID: 44921,
	PackageBase: "cower", Version: "14-2", Description: "A simple AUR agent with a pretentious name",
	URL: "http://github.com/falconindy/cower", NumVotes: 590, Popularity: 24.595536, OutOfDate: 0,
	Maintainer: "falconindy", FirstSubmitted: 1293676237, LastModified: 1441804093,
	URLPath: "/cgit/aur.git/snapshot/cower.tar.gz", Depends: []string{"curl", "openssl", "pacman", "yajl"},
	MakeDepends: []string{"perl"}, CheckDepends: []string(nil), Conflicts: []string(nil),
	Provides: []string(nil), Replaces: []string(nil), OptDepends: []string(nil),
	Groups: []string(nil), License: []string{"MIT"}, Keywords: []string{}}}

func TestNewClient(t *testing.T) {
	newHTTPClient := &http.Client{}

	customRequestEditor := func(ctx context.Context, req *http.Request) error {
		return nil
	}

	type args struct {
		opts []ClientOption
	}
	tests := []struct {
		name             string
		args             args
		wantBaseURL      string
		wanthttpClient   *http.Client
		wantRequestDoers []RequestEditorFn
		wantErr          bool
	}{
		{
			name:             "default",
			args:             args{opts: []ClientOption{}},
			wantBaseURL:      "https://aur.archlinux.org/rpc.php?",
			wanthttpClient:   http.DefaultClient,
			wantErr:          false,
			wantRequestDoers: []RequestEditorFn{},
		},
		{
			name:             "custom base url",
			args:             args{opts: []ClientOption{WithBaseURL("localhost:8000")}},
			wantBaseURL:      "localhost:8000/rpc.php?",
			wanthttpClient:   http.DefaultClient,
			wantErr:          false,
			wantRequestDoers: []RequestEditorFn{},
		},
		{
			name:             "custom base url complete",
			args:             args{opts: []ClientOption{WithBaseURL("localhost:8000/rpc.php?")}},
			wantBaseURL:      "localhost:8000/rpc.php?",
			wanthttpClient:   http.DefaultClient,
			wantErr:          false,
			wantRequestDoers: []RequestEditorFn{},
		},
		{
			name:             "custom http client",
			args:             args{opts: []ClientOption{WithHTTPClient(newHTTPClient)}},
			wantBaseURL:      "https://aur.archlinux.org/rpc.php?",
			wanthttpClient:   newHTTPClient,
			wantErr:          false,
			wantRequestDoers: []RequestEditorFn{},
		},
		{
			name:             "custom request editor",
			args:             args{opts: []ClientOption{WithRequestEditorFn(customRequestEditor)}},
			wantBaseURL:      "https://aur.archlinux.org/rpc.php?",
			wanthttpClient:   newHTTPClient,
			wantErr:          false,
			wantRequestDoers: []RequestEditorFn{customRequestEditor},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantBaseURL, got.BaseURL)
			assert.Equal(t, tt.wanthttpClient, got.HTTPClient)
			assert.Equal(t, len(tt.wantRequestDoers), len(got.RequestEditors))
		})
	}
}

func Test_newAURRPCRequest(t *testing.T) {
	values := url.Values{}
	values.Set("type", "search")
	values.Set("arg", "test-query")
	got, err := newAURRPCRequest(context.Background(), _defaultURL, values)
	assert.NoError(t, err)
	assert.Equal(t, "https://aur.archlinux.org/rpc.php?arg=test-query&type=search&v=5", got.URL.String())
}

func Test_parseRPCResponse(t *testing.T) {
	type args struct {
		resp *http.Response
	}
	tests := []struct {
		name       string
		args       args
		want       []Pkg
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "service unavailable",
			args: args{resp: &http.Response{
				StatusCode: 503,
				Body:       ioutil.NopCloser(bytes.NewBufferString("{}"))}},
			want:       []Pkg{},
			wantErr:    true,
			wantErrMsg: "AUR is unavailable at this moment",
		},
		{
			name: "ok empty body",
			args: args{resp: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString("{}"))}},
			want:       nil,
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "ok empty body",
			args: args{resp: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString("{}"))}},
			want:       nil,
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "payload error",
			args: args{resp: &http.Response{
				StatusCode: 400,
				Body:       ioutil.NopCloser(bytes.NewBufferString(errorPayload))}},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "status 400: Incorrect by field specified.",
		},
		{
			name: "valid payload",
			args: args{resp: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(validPayload))}},
			want:       validPayloadItems,
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRPCResponse(tt.args.resp)

			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)

				return
			} else {
				assert.NoError(t, err)
			}

			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestClient_applyEditors_client(t *testing.T) {
	requestEditor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("test", "value-test")
		return nil
	}
	c := &Client{
		BaseURL:        "https://aur.archlinux.org/rpc.php?",
		HTTPClient:     http.DefaultClient,
		RequestEditors: []RequestEditorFn{requestEditor},
	}

	req := &http.Request{Header: http.Header{}}

	err := c.applyEditors(context.Background(), req, []RequestEditorFn{})

	assert.NoError(t, err)
	assert.Equal(t, "value-test", req.Header.Get("test"))
}

func TestClient_applyEditors_extra(t *testing.T) {
	requestEditor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("test", "value-test")
		return nil
	}
	c := &Client{
		BaseURL:        "https://aur.archlinux.org/rpc.php?",
		HTTPClient:     http.DefaultClient,
		RequestEditors: []RequestEditorFn{},
	}

	req := &http.Request{Header: http.Header{}}

	err := c.applyEditors(context.Background(), req, []RequestEditorFn{requestEditor})

	assert.NoError(t, err)
	assert.Equal(t, "value-test", req.Header.Get("test"))
}

func TestClient_applyEditors_error(t *testing.T) {
	requestEditor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("test", "value-test")
		return ErrServiceUnavailable
	}
	c := &Client{
		BaseURL:        "https://aur.archlinux.org/rpc.php?",
		HTTPClient:     http.DefaultClient,
		RequestEditors: []RequestEditorFn{},
	}

	req := &http.Request{Header: http.Header{}}

	err := c.applyEditors(context.Background(), req, []RequestEditorFn{requestEditor})
	assert.Error(t, err)
}

type MockedClient struct {
	mock.Mock
}

func (m *MockedClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)

	return args.Get(0).(*http.Response), args.Error(1)

}

func TestClient_Search(t *testing.T) {
	testClient := new(MockedClient)

	c := &Client{
		BaseURL:        "https://aur.archlinux.org/rpc.php?",
		HTTPClient:     testClient,
		RequestEditors: []RequestEditorFn{},
	}

	testClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString(validPayload))}, nil)

	got, err := c.Search(context.Background(), "test", Name)

	assert.NoError(t, err)

	assert.Equal(t, validPayloadItems, got)

	testClient.AssertNumberOfCalls(t, "Do", 1)
	testClient.AssertExpectations(t)

	requestMade := testClient.Calls[0].Arguments.Get(0).(*http.Request)
	assert.Equal(t, "https://aur.archlinux.org/rpc.php?arg=test&by=name&type=search&v=5",
		requestMade.URL.String())
}

func TestClient_Info(t *testing.T) {
	testClient := new(MockedClient)

	c := &Client{
		BaseURL:        "https://aur.archlinux.org/rpc.php?",
		HTTPClient:     testClient,
		RequestEditors: []RequestEditorFn{},
	}

	testClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString(validPayload))}, nil)

	got, err := c.Info(context.Background(), []string{"test"})

	assert.NoError(t, err)

	assert.Equal(t, validPayloadItems, got)

	testClient.AssertNumberOfCalls(t, "Do", 1)
	testClient.AssertExpectations(t)

	requestMade := testClient.Calls[0].Arguments.Get(0).(*http.Request)
	assert.Equal(t, "https://aur.archlinux.org/rpc.php?arg%5B%5D=test&type=info&v=5",
		requestMade.URL.String())
}

func TestClient_InfoNoMatch(t *testing.T) {
	testClient := new(MockedClient)

	c := &Client{
		BaseURL:        "https://aur.archlinux.org/rpc.php?",
		HTTPClient:     testClient,
		RequestEditors: []RequestEditorFn{},
	}

	testClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString(noMatchPayload))}, nil)

	got, err := c.Info(context.Background(), []string{"test"})

	assert.NoError(t, err)

	assert.Equal(t, []Pkg{}, got)

	testClient.AssertNumberOfCalls(t, "Do", 1)
	testClient.AssertExpectations(t)

	requestMade := testClient.Calls[0].Arguments.Get(0).(*http.Request)
	assert.Equal(t, "https://aur.archlinux.org/rpc.php?arg%5B%5D=test&type=info&v=5",
		requestMade.URL.String())
}

func TestClient_InfoError(t *testing.T) {
	testClient := new(MockedClient)

	c := &Client{
		BaseURL:        "https://aur.archlinux.org/rpc.php?",
		HTTPClient:     testClient,
		RequestEditors: []RequestEditorFn{},
	}

	testClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 503,
		Body:       ioutil.NopCloser(bytes.NewBufferString(errorPayload))}, nil)

	_, err := c.Info(context.Background(), []string{"test"})

	assert.ErrorIs(t, ErrServiceUnavailable, err)

	testClient.AssertNumberOfCalls(t, "Do", 1)
	testClient.AssertExpectations(t)

	requestMade := testClient.Calls[0].Arguments.Get(0).(*http.Request)
	assert.Equal(t, "https://aur.archlinux.org/rpc.php?arg%5B%5D=test&type=info&v=5",
		requestMade.URL.String())
}
