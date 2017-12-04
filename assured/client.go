package assured

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	kitlog "github.com/go-kit/kit/log"
)

// Client
type Client struct {
	Errc       chan error
	Port       int
	logger     kitlog.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	httpClient *http.Client
}

// NewDefaultClient creates a new go-rest-assured client with default parameters
func NewDefaultClient() *Client {
	return NewClient(nil, 0, nil)
}

// NewClient creates a new go-rest-assured client
func NewClient(root context.Context, port int, logger *kitlog.Logger) *Client {
	if root == nil {
		root = context.Background()
	}
	if logger == nil {
		l := kitlog.NewLogfmtLogger(ioutil.Discard)
		logger = &l
	}
	if port == 0 {
		if listen, err := net.Listen("tcp", ":0"); err == nil {
			port = listen.Addr().(*net.TCPAddr).Port
			listen.Close()
		}
	}
	ctx, cancel := context.WithCancel(root)
	c := Client{
		Errc:       make(chan error),
		logger:     *logger,
		Port:       port,
		ctx:        ctx,
		cancel:     cancel,
		httpClient: &http.Client{},
	}
	StartApplicationHTTPListener(c.ctx, c.logger, c.Port, c.Errc)
	return &c
}

// URL returns the url to use to test you stubbed endpoints
func (c *Client) URL() string {
	return fmt.Sprintf("http://localhost:%d/when", c.Port)
}

// Close is used to close the running service
func (c *Client) Close() {
	c.cancel()
}

// Given stubs an assured Call
func (c *Client) Given(call Call) error {
	var req *http.Request
	var err error

	if call.Method == "" {
		return fmt.Errorf("cannot stub call without Method")
	}

	if call.Response == nil {
		req, err = http.NewRequest(call.Method, fmt.Sprintf("http://localhost:%d/given/%s", c.Port, call.Path), nil)
	} else {
		req, err = http.NewRequest(call.Method, fmt.Sprintf("http://localhost:%d/given/%s", c.Port, call.Path), bytes.NewReader(call.Response))
	}
	if err != nil {
		return err
	}
	if call.StatusCode != 0 {
		req.Header.Set("Assured-Status", fmt.Sprintf("%d", call.StatusCode))
	}

	_, err = c.httpClient.Do(req)
	return err
}

// Verify returns all of the calls made against a stubbed method and path
func (c *Client) Verify(method, path string) ([]Call, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("http://localhost:%d/verify/%s", c.Port, path), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failure to verify calls")
	}
	defer resp.Body.Close()

	var calls []Call
	if err = json.NewDecoder(resp.Body).Decode(&calls); err != nil {
		return nil, err
	}
	return calls, nil
}

// Clear assured calls for a Method and Path
func (c *Client) Clear(method, path string) error {
	req, err := http.NewRequest(method, fmt.Sprintf("http://localhost:%d/clear/%s", c.Port, path), nil)
	if err != nil {
		return err
	}
	_, err = c.httpClient.Do(req)
	return err
}

// ClearAll clears all assured calls
func (c *Client) ClearAll() error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost:%d/clear", c.Port), nil)
	if err != nil {
		return err
	}
	_, err = c.httpClient.Do(req)
	return err
}
