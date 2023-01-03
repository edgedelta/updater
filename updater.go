package updater

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/edgedelta/updater/api"
	"github.com/edgedelta/updater/core"
	"github.com/edgedelta/updater/k8s"
	"github.com/edgedelta/updater/log"

	"github.com/go-yaml/yaml"
	"k8s.io/client-go/rest"
)

var (
	confVarRe                        = regexp.MustCompile(`{{\s*([^{} ]+)\s*}}`)
	contextualVariableTemplateFormat = `{{ index .Vars "%s" }}`
)

type Updater struct {
	config *core.UpdaterConfig
	apiCli core.VersioningServiceClient

	k8sCliOpts []k8s.NewClientOpt
	k8sCli     *k8s.Client
}

type NewClientOpt func(*Updater)

func WithK8sConfig(config *rest.Config) NewClientOpt {
	return func(u *Updater) {
		u.k8sCliOpts = append(u.k8sCliOpts, k8s.WithConfig(config))
	}
}

func WithConfig(config *core.UpdaterConfig) NewClientOpt {
	return func(u *Updater) {
		u.config = config
	}
}

func NewUpdater(ctx context.Context, configPath string, opts ...NewClientOpt) (*Updater, error) {
	u := &Updater{k8sCliOpts: make([]k8s.NewClientOpt, 0)}
	for _, o := range opts {
		o(u)
	}
	if u.config == nil {
		u.config = &core.UpdaterConfig{}
		b, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(b, u.config); err != nil {
			return nil, err
		}
	}
	cl, err := k8s.NewClient(u.k8sCliOpts...)
	if err != nil {
		return nil, err
	}
	u.k8sCli = cl
	if err := u.evaluateConfigVars(ctx); err != nil {
		return nil, fmt.Errorf("updater.Updater.evaluateConfigVars: %v", err)
	}
	if err := u.validateEntities(); err != nil {
		return nil, fmt.Errorf("updater.Updater.validateEntities: %v", err)
	}
	u.apiCli = api.NewClient(&u.config.API)
	if u.config.API.MetadataEndpoint != nil {
		u.config.Metadata, err = u.apiCli.GetMetadata()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch metadata, err: %v", err)
		}
	}
	if err := u.evaluateMetadataConfigVars(); err != nil {
		return nil, fmt.Errorf("updater.Updater.evaluateMetadataConfigVars: %v", err)
	}
	return u, nil
}

func (u *Updater) APIClient() *api.Client {
	return u.apiCli.(*api.Client)
}

func (u *Updater) LogCustomTags() map[string]string {
	m := make(map[string]string)
	for k, v := range u.config.Metadata {
		m[k] = v
	}
	return m
}

func (u *Updater) LogUploaderEnabled() bool {
	return u.config.API.LogUpload != nil && u.config.API.LogUpload.Enabled
}

// validateEntities function validates the given entities through the rules:
//   - Each entity ID is unique
func (u *Updater) validateEntities() error {
	if len(u.config.Entities) == 0 {
		return errors.New("no entity is defined, need at least 1")
	}
	ids := make(map[string]bool)
	for _, e := range u.config.Entities {
		if _, ok := ids[e.ID]; ok {
			return fmt.Errorf("entity ID %s is used at least twice", e.ID)
		}
		ids[e.ID] = true
	}
	return nil
}

func (u *Updater) Run(ctx context.Context) error {
	u.logRunningConfig()
	errors := core.NewErrors()
	for _, entity := range u.config.Entities {
		res, err := u.apiCli.GetLatestApplicableTag(entity.ID)
		if err != nil {
			errors.Addf("failed to get latest applicable tag from API for entity with ID %s, err: %v", entity.ID, err)
			continue
		}
		log.Info("Latest applicable tag from API: %+v", res)
		for _, path := range entity.K8sPaths {
			if err := u.k8sCli.SetResourceKeyValue(ctx, path, res.URL); err != nil {
				errors.Addf("failed to set K8s resource spec key/value for entity with ID %s (path: %s, value: %s), err: %v", entity.ID, path, res.URL, err)
				continue
			}
		}
	}
	return errors.ErrorOrNil()
}

func (u *Updater) evaluateConfigVars(ctx context.Context) (err error) {
	for index, entity := range u.config.Entities {
		if u.config.Entities[index].ID, err = u.evaluateConfigVar(ctx, entity.ID); err != nil {
			return
		}
	}
	if u.config.API.BaseURL, err = u.evaluateConfigVar(ctx, u.config.API.BaseURL); err != nil {
		return
	}
	if u.config.API.TopLevelAuth != nil {
		if u.config.API.TopLevelAuth.HeaderValue, err = u.evaluateConfigVar(ctx, u.config.API.TopLevelAuth.HeaderValue); err != nil {
			return
		}
	}
	if u.config.API.LatestTagEndpoint.Endpoint, err = u.evaluateConfigVar(ctx, u.config.API.LatestTagEndpoint.Endpoint); err != nil {
		return
	}
	if u.config.API.LatestTagEndpoint.Params != nil {
		for k, v := range u.config.API.LatestTagEndpoint.Params.QueryParams {
			if u.config.API.LatestTagEndpoint.Params.QueryParams[k], err = u.evaluateConfigVar(ctx, v); err != nil {
				return
			}
		}
	}
	if u.config.API.LogUpload.PresignedUploadURLEndpoint.Endpoint, err = u.evaluateConfigVar(ctx, u.config.API.LogUpload.PresignedUploadURLEndpoint.Endpoint); err != nil {
		return
	}
	if u.config.API.LogUpload.PresignedUploadURLEndpoint.Params != nil {
		for k, v := range u.config.API.LogUpload.PresignedUploadURLEndpoint.Params.QueryParams {
			if u.config.API.LogUpload.PresignedUploadURLEndpoint.Params.QueryParams[k], err = u.evaluateConfigVar(ctx, v); err != nil {
				return
			}
		}
	}
	if u.config.Log != nil {
		for k, v := range u.config.Log.CustomTags {
			if u.config.Log.CustomTags[k], err = u.evaluateConfigVar(ctx, v); err != nil {
				return
			}
		}
	}
	return nil
}

func (u *Updater) evaluateConfigVar(ctx context.Context, val string) (string, error) {
	var err error
	return confVarRe.ReplaceAllStringFunc(val, func(s string) string {
		inner := confVarRe.FindStringSubmatch(s)[1]
		if strings.HasPrefix(inner, ".k8s.secrets.") {
			path := inner[13:] // .k8s.secrets.<KEY>
			elms := strings.Split(path, ".")
			if len(elms) != 2 {
				err = fmt.Errorf("path should have pattern: .k8s.<NAMESPACE>.<SECRET-NAME>, got '%s' instead", path)
				return ""
			}
			namespace := elms[0]
			name := elms[1]
			var secret string
			secret, err = u.k8sCli.GetSecret(ctx, namespace, name)
			return secret
		}
		if strings.HasPrefix(inner, ".env.") {
			key := inner[5:] // .env.<KEY>
			return os.Getenv(key)
		}
		if strings.HasPrefix(inner, ".ctx.") {
			key := inner[5:] // .ctx.<KEY>

			// Replace the contextual variable's key to a Go template map index key to later
			// use inside the related function(s).
			return fmt.Sprintf(contextualVariableTemplateFormat, key)
		}
		return s // If unmatching, just return itself
	}), err
}

func (u *Updater) evaluateMetadataConfigVars() (err error) {
	if u.config.Log != nil {
		for k, v := range u.config.Log.CustomTags {
			if u.config.Log.CustomTags[k], err = u.evaluateMetadataConfigVar(v); err != nil {
				return
			}
		}
	}
	return nil
}

func (u *Updater) evaluateMetadataConfigVar(val string) (string, error) {
	var err error
	return confVarRe.ReplaceAllStringFunc(val, func(s string) string {
		inner := confVarRe.FindStringSubmatch(s)[1]
		if strings.HasPrefix(inner, ".meta.") {
			key := inner[6:] // .meta.<KEY>
			v, ok := u.config.Metadata[key]
			if !ok {
				err = fmt.Errorf("metadata with key %q is not found", key)
				return ""
			}
			return v
		}
		return s // If unmatching, just return itself
	}), err
}

func (u *Updater) logRunningConfig() {
	entities := make([]string, 0)
	for _, e := range u.config.Entities {
		entities = append(entities, fmt.Sprintf("%s:%s", e.ImageName, e.ID))
	}
	l := fmt.Sprintf("Updater is running for entities %s with API base URL: %s, latest tag endpoint: %s, log uploader is", strings.Join(entities, ", "), u.config.API.BaseURL, u.config.API.LatestTagEndpoint.Endpoint)
	if u.LogUploaderEnabled() {
		l += fmt.Sprintf(" enabled with presigned URL endpoint: %s, encoding: %s, and compression: %s", u.config.API.LogUpload.PresignedUploadURLEndpoint.Endpoint, u.config.API.LogUpload.Encoding.Type, u.config.API.LogUpload.Compression)
	} else {
		l += " disabled"
	}
	log.Debug(l)
}
