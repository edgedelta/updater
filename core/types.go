package core

type UpdaterConfig struct {
	Entities []EntityProperties `yaml:"entities"`
	API      APIConfig          `yaml:"api"`
}

type EntityProperties struct {
	ID        string            `yaml:"id"`
	ImageName string            `yaml:"image"`
	K8sPaths  []K8sResourcePath `yaml:"paths"`
}

type APIConfig struct {
	BaseURL           string         `yaml:"base_url"`
	LatestTagEndpoint EndpointConfig `yaml:"latest_tag"`
	TopLevelAuth      *APIAuth       `yaml:"auth,omitempty"`
}

type EndpointConfig struct {
	Endpoint string     `yaml:"endpoint"`
	Auth     *APIAuth   `yaml:"auth,omitempty"`
	Params   *ParamConf `yaml:"params,omitempty"`
}

type APIAuth struct {
	HeaderKey   string `yaml:"header_key"`
	HeaderValue string `yaml:"header_value"`
}

type ParamConf struct {
	QueryParams map[string]string `yaml:"query,omitempty"`
	PathParams  map[string]string `yaml:"path,omitempty"`
}

type LatestTagResponse struct {
	Tag   string `json:"tag"`
	Image string `json:"image"`
	URL   string `json:"url"`
}

type VersioningServiceClient interface {
	GetLatestApplicableTag(entityID string) (*LatestTagResponse, error)
}
