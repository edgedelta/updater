package core

type UpdaterConfig struct {
	Entities []EntityProperties `yaml:"entities"`
	API      APIConfig          `yaml:"api"`
	Log      *LogConfig         `yaml:"log,omitempty"`
	Metadata map[string]string  `yaml:"-"`
}

type EntityProperties struct {
	ID        string            `yaml:"id"`
	ImageName string            `yaml:"image"`
	K8sPaths  []K8sResourcePath `yaml:"paths"`
}

type APIConfig struct {
	BaseURL           string           `yaml:"base_url"`
	LatestTagEndpoint EndpointConfig   `yaml:"latest_tag"`
	MetadataEndpoint  *EndpointConfig  `yaml:"metadata,omitempty"`
	LogUpload         *LogUploadConfig `yaml:"log_upload,omitempty"`
	TopLevelAuth      *APIAuth         `yaml:"auth,omitempty"`
}

type LogConfig struct {
	CustomTags map[string]string `yaml:"custom_tags,omitempty"`
}

type LogUploadConfig struct {
	Enabled                    bool            `yaml:"enabled"`
	PresignedUploadURLEndpoint EndpointConfig  `yaml:"presigned_upload_url"`
	Method                     string          `yaml:"method"`
	Encoding                   *EncodingConfig `yaml:"encoding"`
	Compression                CompressionType `yaml:"compression,omitempty"`
	Params                     *ParamConf      `yaml:"params,omitempty"`
}

type EndpointConfig struct {
	Endpoint string     `yaml:"endpoint"`
	Params   *ParamConf `yaml:"params,omitempty"`
}

type APIAuth struct {
	HeaderKey   string `yaml:"header_key"`
	HeaderValue string `yaml:"header_value"`
}

type ParamConf struct {
	QueryParams map[string]string `yaml:"query,omitempty"`
}

type EncodingConfig struct {
	Type EncodingType     `yaml:"type"`
	Opts *EncodingOptions `yaml:"options,omitempty"`
}

type EncodingType string

const (
	EncodingJSON EncodingType = "json"
	EncodingRaw               = "raw"
)

type EncodingOptions struct {
	Delimiter string `yaml:"delimiter,omitempty"`
}

type CompressionType string

const (
	CompressionGzip CompressionType = "gzip"
	CompressionNoOp                 = ""
)

type LatestTagResponse struct {
	Tag   string `json:"tag"`
	Image string `json:"image"`
	URL   string `json:"url"`
}

type VersioningServiceClient interface {
	GetLatestApplicableTag(entityID, entityName string) (*LatestTagResponse, error)
	GetPresignedLogUploadURL(logSize int) (string, error)
	UploadLogs(lines []any) error
	GetMetadata() (map[string]string, error)
}
