package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/edgedelta/updater/core"
	"github.com/edgedelta/updater/core/compressors"
	"github.com/edgedelta/updater/core/encoders"
	"github.com/edgedelta/updater/log"
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
		c.conf.LatestTagEndpoint.Params, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("constructURLWithParams err: %v", err)
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

func (c *Client) GetPresignedLogUploadURL(logSize int) (string, error) {
	log.Debug("api.Client.GetPresignedLogUploadURL: Called with log size %d", logSize)
	url, err := constructURLWithParams(
		c.conf.BaseURL+c.conf.LogUpload.PresignedUploadURLEndpoint.Endpoint,
		c.conf.LogUpload.PresignedUploadURLEndpoint.Params, map[string]string{
			"size": fmt.Sprintf("%d", logSize),
		},
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

func (c *Client) UploadLogs(lines []interface{}) error {
	wr := new(bytes.Buffer)
	comp, err := compressors.New(wr, c.conf.LogUpload.Compression)
	if err != nil {
		return fmt.Errorf("compressors.New: %v", err)
	}
	enc, err := encoders.New(comp, c.conf.LogUpload.Encoding)
	if err != nil {
		return fmt.Errorf("encoders.New: %v", err)
	}
	if err := enc.Write(lines); err != nil {
		return fmt.Errorf("encoders.Encoder.Write: %v", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("encoders.Encoder.Close: %v", err)
	}
	if err := comp.Flush(); err != nil {
		return fmt.Errorf("compressors.Compressor.Flush: %v", err)
	}
	if err := comp.Close(); err != nil {
		return fmt.Errorf("compressors.Compressor.Close: %v", err)
	}
	presignedURL, err := c.GetPresignedLogUploadURL(wr.Len())
	if err != nil {
		return fmt.Errorf("api.Client.GetPresignedLogUploadURL: %v", err)
	}
	url, err := constructURLWithParams(presignedURL, c.conf.LogUpload.Params, nil)
	if err != nil {
		return fmt.Errorf("constructURLWithParams: %v", err)
	}
	req, err := http.NewRequest(c.conf.LogUpload.Method, url, wr)
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

func constructURLWithParams(base string, params *core.ParamConf, ctxVars map[string]string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("url.Parse: %v", err)
	}
	if params == nil {
		return u.String(), nil
	}
	q := u.Query()
	for k, v := range params.QueryParams {
		val, err := core.EvaluateContextualTemplate(v, ctxVars)
		if err != nil {
			return "", fmt.Errorf("core.EvaluateContextualTemplate: %v", err)
		}
		q.Add(k, val)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}
