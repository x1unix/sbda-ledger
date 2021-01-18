package ledger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const msgOk = "pong"

type pingResponse struct {
	Message string `json:"message"`
}

type Client struct {
	http    *http.Client
	baseUrl string
}

func NewClient(h *http.Client, baseUrl string) *Client {
	return &Client{http: h, baseUrl: baseUrl}
}

func (c Client) newRequest(method, reqPath string, data interface{}, auth Token) (*http.Request, error) {
	var body io.Reader
	if data != nil {
		data, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare request: %w", err)
		}

		body = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, c.baseUrl+reqPath, body)
	if err != nil {
		return nil, fmt.Errorf("can't prepare request: %w", err)
	}

	if auth != "" {
		auth.apply(req)
	}

	return req, nil
}

func (c Client) do(req *http.Request, out interface{}) error {
	rsp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	defer rsp.Body.Close()
	content, err := ioutil.ReadAll(rsp.Body)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read response: %w", err)
	}

	switch rsp.StatusCode {
	case http.StatusOK:
		if out == nil {
			return fmt.Errorf("got response but passed output is nil")
		}

		return json.Unmarshal(content, out)
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound, http.StatusBadGateway:
		return errors.New(rsp.Status)
	default:
		errRsp := ErrorResponse{Status: rsp.Status}
		if err := json.Unmarshal(content, &errRsp); err != nil {
			errRsp.ErrorData.Error = string(content)
		}

		return errRsp
	}
}

func (c Client) post(reqPath string, data interface{}, out interface{}, auth Token) error {
	req, err := c.newRequest(http.MethodPost, reqPath, data, auth)
	if err != nil {
		return err
	}

	return c.do(req, out)
}

func (c Client) get(reqPath string, out interface{}, auth Token) error {
	req, err := c.newRequest(http.MethodGet, reqPath, nil, auth)
	if err != nil {
		return err
	}

	return c.do(req, out)
}

func (c Client) delete(reqPath string, auth Token) error {
	req, err := c.newRequest(http.MethodGet, reqPath, nil, auth)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c Client) Ping() error {
	out := new(pingResponse)
	if err := c.get("/ping", out, ""); err != nil {
		return err
	}

	if out.Message != msgOk {
		return fmt.Errorf("unexpected ping contents: %q", out.Message)
	}
	return nil
}
