# ORT Operator API

This repository contains a HTTP/JSON API, a Matrix chatbot and a (currently broken) Slack chatbot for the
[ORT Kubernetes Operator](https://github.com/haikoschol/ort-operator). The bots should probably live in separate repos.

## HTTP Endpoints

### `GET /runs` - Returns a list of all OrtRun resources

Response:

```json
{
  "runs": [
    {
      "name": "<name>",
      "repoUrl": "<repoUrl>",
      "status": {
        "analyzer": "[Pending|Running|Succeeded|Failed|Aborted]",
        "scanner": "[Pending|Running|Succeeded|Failed|Aborted]",
        "reporter": "[Pending|Running|Succeeded|Failed|Aborted]"
      }
    }
  ]
}
```

### `GET /runs/<name>` - Return the OrtRun resources with the given name

Response:

```json
{
  "name": "<name>",
  "repoUrl": "<repoUrl>",
  "status": {
    "analyzer": "[Pending|Running|Succeeded|Failed|Aborted]",
    "scanner": "[Pending|Running|Succeeded|Failed|Aborted]",
    "reporter": "[Pending|Running|Succeeded|Failed|Aborted]"
  },
  "kubernetesResource": "<yaml>"
}
```

### `POST /runs` - Create a new OrtRun resources

Payload:

```json
{
    "repoUrl": "https://github.com/haikoschol/cats-of-asia.git"
}
```

Response:

```json
{
  "name": "<name>",
  "repoUrl": "<repoUrl>",
  "status": {
    "analyzer": "[Pending|Running|Succeeded|Failed|Aborted]",
    "scanner": "[Pending|Running|Succeeded|Failed|Aborted]",
    "reporter": "[Pending|Running|Succeeded|Failed|Aborted]"
  },
  "kubernetesResource": "<yaml>"
}
```

### `GET /logs/<name>/[analyzer|scanner|reporter]` - Fetch the logs from the analyzer, scanner or reporter of an ORT run

Response:

```json
{
  "name": "<name>",
  "stage": "[analyzer|scanner|reporter]",
  "podLogs": [
    {
      "podName": "<name>",
      "podLogs": "<log>"
    }
  ]
}
```

## Configuration

To talk to Kubernetes, the API process first tries [InClusterConfig](https://pkg.go.dev/k8s.io/client-go/rest#InClusterConfig)
and if that fails looks for a kubeconfig in `$HOME/.kube/config`.

if `MATRIX_SERVER` is set, `MATRIX_USER` and `MATRIX_ACCESS_TOKEN` are assumed to be set as well and an instance of the
Matrix bot is created and run.
