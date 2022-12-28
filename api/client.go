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
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if c.conf.TopLevelAuth != nil {
		req.Header.Add(c.conf.TopLevelAuth.HeaderKey, c.conf.TopLevelAuth.HeaderValue)
	}
	if c.conf.LatestTagEndpoint.Auth != nil {
		req.Header.Add(c.conf.LatestTagEndpoint.Auth.HeaderKey, c.conf.LatestTagEndpoint.Auth.HeaderValue)
	}
	res, err := c.cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("StatusCode != 200 (%d), response body: %s", res.StatusCode, string(data))
	}
	var r core.LatestTagResponse
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return &r, nil
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
