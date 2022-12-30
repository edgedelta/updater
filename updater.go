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
	confVarRe = regexp.MustCompile(`{{\s*([^{} ]+)\s*}}`)
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
		log.Debug("Config file %s contents: %s", configPath, string(b))
		if err := yaml.Unmarshal(b, u.config); err != nil {
			return nil, err
		}
	}
	cl, err := k8s.NewClient(u.k8sCliOpts...)
	if err != nil {
		return nil, err
	}
	u.k8sCli = cl
	if err := u.EvaluateConfigVars(ctx); err != nil {
		return nil, fmt.Errorf("EvaluateConfigVars: %v", err)
	}
	if err := u.ValidateEntities(); err != nil {
		return nil, fmt.Errorf("ValidateEntities: %v", err)
	}
	u.apiCli = api.NewClient(&u.config.API)
	return u, nil
}

func (u *Updater) APIClient() *api.Client {
	return u.apiCli.(*api.Client)
}

// ValidateEntities function validates the given entities through the rules:
//   - Each entity ID is unique
func (u *Updater) ValidateEntities() error {
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
	errors := core.NewErrors()
	for _, entity := range u.config.Entities {
		res, err := u.apiCli.GetLatestApplicableTag(entity.ID)
		if err != nil {
			errors.Addf("failed to get latest applicable tag from API for entity with ID %s, err: %v", entity.ID, err)
			continue
		}
		for _, path := range entity.K8sPaths {
			if err := u.k8sCli.SetResourceKeyValue(ctx, path, res.URL); err != nil {
				errors.Addf("failed to set K8s resource spec key/value for entity with ID %s (path: %s, value: %s), err: %v", entity.ID, path, res.URL, err)
				continue
			}
			log.Info("Updated version of resource with path %s to %s", path, res.URL)
		}
	}
	return errors.ErrorOrNil()
}

func (u *Updater) EvaluateConfigVars(ctx context.Context) (err error) {
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
		for k, v := range u.config.API.LatestTagEndpoint.Params.PathParams {
			if u.config.API.LatestTagEndpoint.Params.QueryParams[k], err = u.evaluateConfigVar(ctx, v); err != nil {
				return
			}
		}
	}
	if u.config.API.LatestTagEndpoint.Auth != nil {
		if u.config.API.LatestTagEndpoint.Auth.HeaderKey, err = u.evaluateConfigVar(ctx, u.config.API.LatestTagEndpoint.Auth.HeaderKey); err != nil {
			return
		}
		if u.config.API.LatestTagEndpoint.Auth.HeaderValue, err = u.evaluateConfigVar(ctx, u.config.API.LatestTagEndpoint.Auth.HeaderValue); err != nil {
			return
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
			}
			namespace := elms[0]
			name := elms[1]
			var secret string
			secret, err = u.k8sCli.GetSecret(ctx, namespace, name)
			return secret
		}
		if strings.HasPrefix(inner, ".env") {
			key := inner[5:] // .env.<KEY>
			return os.Getenv(key)
		}
		err = fmt.Errorf("config var '%s' starts with an unknown item '%s' (expected '.k8s' or '.env')", s, inner)
		return ""
	}), err
}
