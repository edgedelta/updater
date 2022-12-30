package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/edgedelta/updater/core"
)

type Client struct {
	cl   *http.Client
	conf *core.APIConfig
}

func NewClient(conf *core.APIConfig) *Client {
	if conf == nil {
		return nil
	}
	cl := http.DefaultClient
	return &Client{cl: cl, conf: conf}
}

func (c *Client) GetLatestApplicableTag(id string) (*core.LatestTagResponse, error) {
	url, err := constructURLWithParams(
		c.conf.BaseURL+c.conf.LatestTagEndpoint.Endpoint,
		c.conf.LatestTagEndpoint.Params,
	)
	if err != nil {
		return nil, fmt.Errorf("constructURLWithParams: %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %v", err)
	}
	if c.conf.TopLevelAuth != nil {
		req.Header.Add(c.conf.TopLevelAuth.HeaderKey, c.conf.TopLevelAuth.HeaderValue)
	}
	if c.conf.LatestTagEndpoint.Auth != nil {
		req.Header.Add(c.conf.LatestTagEndpoint.Auth.HeaderKey, c.conf.LatestTagEndpoint.Auth.HeaderValue)
	}
	res, err := c.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http.Client.Do: %v", err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %v", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("status code is not in the expected range (%d), response body: %q", res.StatusCode, string(data))
	}
	var r core.LatestTagResponse
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("json.Unmarshall: %v", err)
	}
	return &r, nil
}

func (c *Client) GetPresignedLogUploadURL() (string, error) {
	url, err := constructURLWithParams(
		c.conf.BaseURL+c.conf.LogUpload.PresignedUploadURLEndpoint.Endpoint,
		c.conf.LogUpload.PresignedUploadURLEndpoint.Params,
	)
	if err != nil {
		return "", fmt.Errorf("constructURLWithParams: %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("http.NewRequest: %v", err)
	}
	if c.conf.TopLevelAuth != nil {
		req.Header.Add(c.conf.TopLevelAuth.HeaderKey, c.conf.TopLevelAuth.HeaderValue)
	}
	if c.conf.LogUpload.PresignedUploadURLEndpoint.Auth != nil {
		req.Header.Add(c.conf.LogUpload.PresignedUploadURLEndpoint.Auth.HeaderKey, c.conf.LogUpload.PresignedUploadURLEndpoint.Auth.HeaderValue)
	}
	res, err := c.cl.Do(req)
	if err != nil {
		return "", fmt.Errorf("http.Client.Do: %v", err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("io.ReadAll: %v", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("status code is not in the expected range (%d), response body: %q", res.StatusCode, string(data))
	}
	var presignedURL string
	if err := json.Unmarshal(data, &presignedURL); err != nil {
		return "", fmt.Errorf("json.Unmarshal: %v", err)
	}
	return presignedURL, nil
}

func (c *Client) UploadLogs(lines []string) error {
	presignedURL, err := c.GetPresignedLogUploadURL()
	if err != nil {
		return fmt.Errorf("api.Client.GetPresignedLogUploadURL: %v", err)
	}
	url, err := constructURLWithParams(presignedURL, c.conf.LogUpload.PresignedUploadURLEndpoint.Params)
	if err != nil {
		return fmt.Errorf("constructURLWithParams: %v", err)
	}
	req, err := http.NewRequest(c.conf.LogUpload.Method, url, nil)
	if err != nil {
		return fmt.Errorf("http.NewRequest: %v", err)
	}
	if c.conf.TopLevelAuth != nil {
		req.Header.Add(c.conf.TopLevelAuth.HeaderKey, c.conf.TopLevelAuth.HeaderValue)
	}
	if c.conf.LogUpload.Auth != nil {
		req.Header.Add(c.conf.LogUpload.Auth.HeaderKey, c.conf.LogUpload.Auth.HeaderValue)
	}
	res, err := c.cl.Do(req)
	if err != nil {
		return fmt.Errorf("http.Client.Do: %v", err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %v", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("status code is not in the expected range (%d), response body: %q", res.StatusCode, string(data))
	}
	return nil
}

func constructURLWithParams(base string, params *core.ParamConf) (string, error) {
	if params == nil {
		return base, nil
	}

	t, err := template.New("url-template").Parse(base)
	if err != nil {
		return "", err
	}

	// TODO: This will lead to an ugly templating syntax, change it to on-demand struct fields
	data := map[string]map[string]string{
		"path":  params.PathParams,
		"query": params.QueryParams,
	}

	buffer := new(bytes.Buffer)
	if err := t.Execute(buffer, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}
