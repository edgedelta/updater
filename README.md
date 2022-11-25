# Edge Delta Agent Updater

## Local deployment

1. Install minikube
2. Install the Edge Delta agent to Minikube
3. Run the test API:

```bash
go run test/api/main.go
```

4. Run the command(s) below to build and deploy OR just deploy the agent updater:

    a. To deploy the latest:
   
    ```bash
    ./deploy/scripts/deploy.sh public.ecr.aws/v4z2v9g0/edgedelta-development:updater-linux-arm64-local "http://host.minikube.internal:8080" "/"
    ```

    b. To build and deploy for local:

    ```bash
    ED_MODE=local ./deploy/scripts/build_and_deploy.sh
    ```

After deployment, you should see a message similar to the one below:

```json
{"level":"info","time":"2022-11-25T10:56:07Z","message":"Updated version of resource with path edgedelta:ds/edgedelta:spec.template.spec.containers[0].image to gcr.io/edgedelta/agent:v0.1.47"}
```