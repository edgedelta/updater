# Edge Delta Agent Updater

Agent updater is a configurable minimal program that helps with updating your Kubernetes resources. It is designed to be used to update daemonset and deployment images, but can be used to update any spec property of these resources.

## Usage

### Configuration

The updater can be configured using a YAML configuration file. The configuration file can be passed to the updater using the `--config` flag. The configuration file defines the resources to be updated, the API to be used to fetch the latest version, and the logging configuration.

#### Entities

Entities are the resources to be updated. The updater supports updating the following Kubernetes resources:

- DaemonSet
- Deployment

Entities are defined under the `entities` list in the configuration file. The following is an example of a configuration file with two entities:

```yaml
entities:
- id: 111-222-333
  image: some-agent
  paths:
  - default:ds/my-agent:spec.template.spec.containers[0].image
- id: 444-555-666
  image: some-other-agent
  paths:
  - default:ds/my-other-agent:spec.template.spec.containers[0].image
```

| Property | Type | Description | Required |
| ---| --- | --- | --- |
| `id` | `string` | Unique ID of the resource | Yes |
| `image` | `string` | Resource's image kind | Yes |
| `paths` | `[]string` | K8s object paths of the properties to be updated | Yes |


#### API

The API is used to fetch the latest version of the resource. The updater supports HTTP REST APIs. The API interface is defined with the following Go interface:

```go
type VersioningServiceClient interface {
	GetLatestApplicableTag(entityID string) (*LatestTagResponse, error)
	GetPresignedLogUploadURL(logSize int) (string, error)
	UploadLogs(lines []any) error
	GetMetadata() (map[string]string, error)
}
```

The API is defined under the `api` section in the configuration file. The following is an example API configuration:

```yaml
api:
  base_url: http://localhost:8080
  auth:
    header_key: Authorization
    header_value: 'Some_token'
  latest_tag:
    endpoint: /latest-version
  metadata:
    endpoint: /metadata
  log_upload:
    enabled: true
    method: PUT
    encoding:
      type: raw
      options:
        delimiter: '\n'
    compression: gzip
    presigned_upload_url:
      endpoint: /log-upload-link
      params:
        query:
          size: '{{ .ctx.size }}'
          format: json
          compression: gzip
```

| Property | Type | Description | Required |
| ---| --- | --- | --- |
| `base_url` | `string` | Base URL of the API | Yes |
| `auth` | `APIAuth` | Authentication configuration | No |
| `latest_tag` | `EndpointConfig` | Configuration for fetching the latest version | Yes |
| `metadata` | `EndpointConfig` | Configuration for fetching the metadata | Yes |
| `log_upload` | `LogUploadConfig` | Configuration for uploading logs | No |

### Installation

The updater can be deployed to a Kubernetes cluster using the latest image from the public Google Container Registry.

1. Replace the `data.updater-config.yml` under the `examples/cronjob.yml` with your configuration file content.
2. Run the following command to deploy the updater:

```bash
kubectl apply -f examples/rbac.yml
kubectl apply -f examples/rolebinding.yml
kubectl apply -f examples/cronjob.yml
```

3. To verify that the updater is working, you can check the logs of the updater pod:

```bash
kubectl logs -f updater-<random-id>
```
