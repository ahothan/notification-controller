# Provider

The `Provider` API defines how events are encoded and the webhook address where they are dispatched.

## Specification

Spec:

```go
type ProviderSpec struct {
	// Type of provider
	// +kubebuilder:validation:Enum=slack;discord;msteams;rocket;generic;github;gitlab;bitbucket;azuredevops;googlechat;webex;sentry;azureeventhub;telegram;lark;matrix;opsgenie
	// +required
	Type string `json:"type"`

	// Alert channel for this provider
	// +optional
	Channel string `json:"channel,omitempty"`

	// Bot username for this provider
	// +optional
	Username string `json:"username,omitempty"`

	// HTTP/S webhook address of this provider
	// +kubebuilder:validation:Pattern="^(http|https)://"
	// +optional
	Address string `json:"address,omitempty"`

	// HTTP/S address of the proxy
	// +kubebuilder:validation:Pattern="^(http|https)://"
	// +optional
	Proxy string `json:"proxy,omitempty"`

	// Secret reference containing the provider details, valid key names are: address, proxy, token, headers (YAML encoded)
	// +optional
	SecretRef *meta.LocalObjectReference `json:"secretRef,omitempty"`

	// CertSecretRef can be given the name of a secret containing
	// a PEM-encoded CA certificate (`caFile`)
	// +optional
	CertSecretRef *meta.LocalObjectReference `json:"certSecretRef,omitempty"`
}
```


Notification providers:

| Provider        | Type          |
| --------------- | ------------- |
| Alertmanager    | alertmanager  |
| Azure Event Hub | azureeventhub |
| Discord         | discord       |
| Generic webhook | generic       |
| Google Chat     | googlechat    |
| Grafana         | grafana       |
| Lark            | lark          |
| Matrix          | matrix        |
| Microsoft Teams | msteams       |
| Opsgenie        | opsgenie      |
| Rocket          | rocket        |
| Sentry          | sentry        |
| Slack           | slack         |
| Telegram        | telegram      |
| WebEx           | webex         |

Git commit status providers:

| Provider     | Type        |
| ------------ | ----------- |
| Azure DevOps | azuredevops |
| Bitbucket    | bitbucket   |
| GitHub       | github      |
| GitLab       | gitlab      |
Status:

```go
// ProviderStatus defines the observed state of Provider
type ProviderStatus struct {
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
}
```

Status condition types:

```go
const (
	// ReadyCondition represents the fact that a given object has passed
	// validation and was acknowledge by the controller.
	ReadyCondition string = "Ready"
)
```

## Example

### Notifications

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Provider
metadata:
  name: slack
  namespace: default
spec:
  type: slack
  channel: general
  # webhook address (ignored if secretRef is specified)
  address: https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
  # HTTP(S) proxy (optional)
  proxy: https://proxy.corp:8080
  # secret containing the webhook address (optional)
  secretRef:
    name: webhook-url
```

Webhook URL secret:

```sh
kubectl create secret generic webhook-url \
--from-literal=address=https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
```

Note that the secret must contain an `address` field.

The provider type can be: `slack`, `msteams`, `rocket`, `discord`, `googlechat`, `webex`, `sentry`,
`telegram`, `lark`, `matrix`, `azureeventhub`, `opsgenie`, `alertmanager`, `grafana` or `generic`.

Some networks need to use an authenticated proxy to access external services. Therefore, the authentication can be stored as a secret to hide parameters like the username and password.

```sh
kubectl create secret generic webhook-url \
--from-literal=address=https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK \
--from-literal=proxy=http://username:password@proxy_url:proxy_port
```

When type `generic` is specified, the notification controller will post the
incoming [event](event.md) in JSON format to the webhook address.

### Generic webhook

The `generic` webhook triggers an HTTP POST request to the provided endpoint.

The `Gotk-Component` header identifies which component this event is coming
from, e.g. `source-controller`, `kustomize-controller`.

```
POST / HTTP/1.1
Host: example.com
Accept-Encoding: gzip
Content-Length: 452
Content-Type: application/json
Gotk-Component: source-controller
User-Agent: Go-http-client/1.1
```

The body of the request looks like this:

```json
{
  "involvedObject": {
    "kind":"GitRepository",
    "namespace":"flux-system",
    "name":"flux-system",
    "uid":"cc4d0095-83f4-4f08-98f2-d2e9f3731fb9",
    "apiVersion":"source.toolkit.fluxcd.io/v1beta2",
    "resourceVersion":"56921",
  },
  "severity":"info",
  "timestamp":"2006-01-02T15:04:05Z",
  "message":"Fetched revision: main/731f7eaddfb6af01cb2173e18f0f75b0ba780ef1",
  "reason":"info",
  "reportingController":"source-controller",
  "reportingInstance":"source-controller-7c7b47f5f-8bhrp",
}
```

The `involvedObject` key contains the object that triggered the event.

You can add additional headers to the POST request by providing a `headers` field to the secret
referenced by the provider. An example is given below:

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Provider
metadata:
  name: generic
  namespace: default
spec:
  type: generic
  address: https://api.github.com/repos/owner/repo/dispatches
  secretRef:
    name: generic-secret
---
apiVersion: v1
kind: Secret
metadata:
  name: generic-secret
  namespace: default
stringData:
  headers: |
     Authorization: token
     X-Forwarded-Proto: https
```

### Self-signed certificates

The `certSecretRef` field names a secret with TLS certificate data. This is for the purpose
of enabling a provider to communicate with a server using a self-signed cert.

To use the field create a secret, containing a CA file, in the same namespace and reference
it from the provider.

```sh
kubectl create secret generic tls-certs \
  --from-file=caFile=ca.crt
```

### Sentry

The sentry provider uses the `channel` field to specify which environment the messages are sent for:

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Provider
metadata:
  name: sentry
  namespace: default
spec:
  type: sentry
  channel: my-cluster-name
  # webhook address (ignored if secretRef is specified)
  address: https://....@sentry.io/12341234
```

The sentry provider also sends traces for events with the severity Info. This can be disabled by setting
the `eventSeverity` field on the related `Alert` Rule to `error`.

### Telegram

For telegram, You can get the token from [the botfather](https://core.telegram.org/bots#6-botfather)
and use `https://api.telegram.org/` as the api url.

```sh
 k create secret generic telegram-token \
 --from-literal=token=<token> \
 --from-literal=address=https://api.telegram.org
```

Also note that `spec.channel` can be a unique identifier for the target chat
or username of the target channel (in the format @channelusername)

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Provider
metadata:
  name: telegram
  namespace: flux-system
spec:
  type: telegram
  channel: "@fluxtest" # or "-1557265138" (channel id)
  secretRef:
    name: telegram-token
```

### Matrix

For Matrix, the address is the homeserver URL and the token is the access token
returned by a call to `/login` or `/register`.

Create a secret:
```
kubectl create secret generic matrix-token \
--from-literal=token=<access-token> \
--from-literal=address=https://matrix.org # replace with if using a different server
```

Then reference the secret in `spec.secretRef`:

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Provider
metadata:
  name: matrix
  namespace: default
spec:
  type: matrix
  channel: "!jezptmDwEeLapMLjOc:matrix.org"
  secretRef:
    name: matrix-token
```

Note that `spec.channel` holds the room id.

### Lark

For sending notifications to Lark, you will have to
[add a bot to the group](https://www.larksuite.com/hc/en-US/articles/360048487736-Bot-Use-bots-in-groups#III.%20How%20to%20configure%20custom%20bots%20in%20a%20group%C2%A0)
and set up a webhook for the bot. This serves as the address field in the secret:

```shell
kubectl create secret generic lark-token \
--from-literal=address=<lark-webhook-url>
```

Then reference the secret in `spec.secretRef`:

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Provider
metadata:
  name: lark
  namespace: default
spec:
  type: lark
  secretRef:
    name: lark-token
```


### Opsgenie

For sending notifications to Opsgenie, you will have to
[add a REST api integration](https://support.atlassian.com/opsgenie/docs/create-a-default-api-integration/)
and setup a api integration for notification provider.

A secret needs to be generated with the api key given by Opsgenie for the integration

```shell
kubectl create secret generic opsgenie-token \
--from-literal=token=<opsgenie-api-key>
```

Then reference the secret in `spec.secretRef`:

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Provider
metadata:
  name: opsgenie
  namespace: default
spec:
  type: opsgenie
  address: https://api.opsgenie.com/v2/alerts
  secretRef:
    name: opsgenie-token
```

### Prometheus Alertmanager

Sends notifications to [alertmanager v2 api](https://github.com/prometheus/alertmanager/blob/main/api/v2/openapi.yaml) if alert manager has basic authentication configured it is recommended to use
secretRef and include the username:password in the address string.

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Provider
metadata:
  name: alertmanager
  namespace: default
spec:
  type: alertmanager
  # webhook address (ignored if secretRef is specified)
  address: https://....@<alertmanager-url>/api/v2/alerts/"
```

When an event is triggered the provider will send a single alert with at least one annotation for alert which is the "message" found for the event.
If a summary is provided in the alert resource an additional "summary" annotation will be added.

The provider will send the following labels for the event.


| Label       | Description                                                                                        |
| ----------- | -------------------------------------------------------------------------------------------------- |
| alertname   | The string Flux followed by the Kind and the reason for the event e.g FluxKustomizationProgressing |
| severity    | The severity of the event (error|info)                                                             |
| timestamp   | The timestamp of the event                                                                         |
| reason      | The machine readable reason for the objects transition into the current status                     |
| kind        | The kind of the involved object associated with the event                                          |
| name        | The name of the involved object associated with the event                                          |
| namespace   | The namespace of the involved object associated with the event                                     |

### Slack App

It is possible to use a Slack App bot integration to send messages. To obtain a bot token, follow
[Slack's guide on bot users](https://api.slack.com/bot-users).

Differences from the Slack [webhook method](#notifications):

* Possible to use single credentials to post to different channels (by adding the integration to each channel)
* All messages are posted with the app username, and not the name of the controller (e.g. `helm-controller, `source-controller`)

To enable the Slack App, the secret must contain the URL of the [chat.postMessage](https://api.slack.com/methods/chat.postMessage)
method and your Slack bot token (starts with `xoxb-`):

```shell
kubectl create secret generic slack-token \
--from-literal=address=https://slack.com/api/chat.postMessage \
--from-literal=token=xoxb-YOUR-TOKEN
```

Then reference this secret in `spec.secretRef`:

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Provider
metadata:
  name: slack
  namespace: default
spec:
  type: slack
  channel: general
  # HTTP(S) proxy (optional)
  proxy: https://proxy.corp:8080
  # secret containing Slack API address and token
  secretRef:
    name: slack-token
```

### Webex App

General steps on how to hook up Flux notifications to a Webex space:

From the Webex App UI:
- create a Webex space where you want notifications to be sent
- after creating a Webex bot (described in next section), add the bot email address to the Webex space ("People | Add people")

Register to https://developer.webex.com/, after signing in:
- create a bot for forwarding FluxCD notifications to a Webex Space (User profile icon | MyWebexApps | Create a New App | Create a Bot)
- make a note of the bot email address, this email needs to be added to the Webex space from the Webex App
- generate a bot access token, this is the ID to use in the kubernetes Secret "token" field (see example below)
- find the room ID associated to the webex space using https://developer.webex.com/docs/api/v1/rooms/list-rooms (select GET, click on "Try It" and search the GET results for the matching Webex space entry), this is the ID to use in the webex Provider manifest "channel" field


Manifests template to use:

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Provider
metadata:
  name: webex
  namespace: flux-system
spec:
  type: webex
  address: https://webexapis.com/v1/messages
  channel: <webexSpaceRoomID>
  secretRef:
    name: webex-bot-access-token
---
apiVersion: v1
kind: Secret
metadata:
  name: webex-bot-access-token
  namespace: flux-system
data:
  # bot access token - must be base64 encoded
  token: <webexBotAccessTokenBase64>
```

Notes:

- spec.address should always be set to the same global Webex API gateway https://webexapis.com/v1/messages
- spec.channel should contain the Webex space room ID as obtained from https://developer.webex.com/ (long alphanumeric string copied as is)
- token in the Secret manifest is the bot access token generated after creating the bot (as for all secrets, must be base64 encoded using for example
"echo -n <token> | base64")

If you do not see any notifications in the targeted Webex space:
- check that you have applied an Alert with the right even sources and providerRef
- check the notification controller log for any error messages
- check that you have added the bot email address to the Webex space, if the bot email address is not added to the space, the notification controller will log a 404 room not found error every time a notification is sent out

Full example of manifests with real looking but fictive room ID and access token:

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Provider
metadata:
  name: webex-fluxcd-space
  namespace: flux-system
spec:
  type: webex
  address: https://webexapis.com/v1/messages
  channel: Y2jzY29zcGFyazovL3VzL1JPT00vMGU3YzZhODAlOWU4MC0xMWVjLWJlZWMtMzNm4DkwQGYwMjIz
  secretRef:
    name: webex-bot-access-token
---
apiVersion: v1
kind: Secret
metadata:
  name: webex-bot-access-token
  namespace: flux-system
data:
  token: TVdaM05UVTFNV1F0WkRBMU55MDKObVkzTFdJek16SXRNems1WVRZM09UVmhNbUprTTJFMk9HVTDaR0l0T1RVNF9QRjg0XzFlYjY1ZmRmLTk2NDMtNDE3Zi05OTc0LWFkNzJjYWUwZTEwZg==
---
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Alert
metadata:
  name: webex-fluxcd-space-alerts
  namespace: flux-system
spec:
  providerRef:
    name: webex-fluxcd-space
  eventSeverity: info
  eventSources:
    - kind: GitRepository
      name: '*'
    - kind: HelmRelease
      name: '*'
    - kind: HelmRepository
      name: '*'
    - kind: Kustomization
      name: '*'
```


### Grafana

To send notifications to [Grafana annotations API](https://grafana.com/docs/grafana/latest/http_api/annotations/),
you have to enable the annotations on a Dashboard like so:

- Annotations > Query > Enable Match any
- Annotations > Query > Tags (Add Tag: `flux`)

If Grafana has authentication configured, create a Kubernetes Secret with the API URL and the API token:
```shell
kubectl create secret generic grafana-token \
--from-literal=token=<grafana-api-key> \
--from-literal=address=https://<grafana-url>/api/annotations
```

Grafana can also use `basic authorization` to authenticate the requests, if both token and
username/password are set in the secret, then `API token` takes precedence over `basic auth`.
```shell
kubectl create secret generic grafana-token \
--from-literal=username=<your-grafana-username> \
--from-literal=password=<your-grafana-password>
```

Then reference the secret in `spec.secretRef`:

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Provider
metadata:
  name: grafana
  namespace: default
spec:
  type: grafana
  secretRef:
    name: grafana-token
```

### Git commit status

The GitHub, GitLab, Bitbucket, and Azure DevOps provider will write to the
commit status in the git repository from which the event originates from.

{{% alert color="info" title="Limitations" %}}
The git notification providers require that a commit hash present in the meta data
of the event. Therefore the the providers will only work with `Kustomization` as an
event source, as it is the only resource which includes this data.
{{% /alert %}}

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Provider
metadata:
  name: podinfo
  namespace: default
spec:
  # provider type can be github or gitlab
  type: github
  address: https://github.com/stefanprodan/podinfo
  secretRef:
    name: api-token
```

The provider type can be: `github`, `gitlab`, `bitbucket` or `azuredevops`.

For bitbucket, the token should contain the username and [app password](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/#Create-an-app-password)
in the format `<username>:<password>`. The app password should have `Repositories (Read/Write)` permission.

You can create the secret using this command:
```shell
kubectl create secret generic api-token --from-literal=token=<username>:<app-password>
```

#### Authentication

GitHub. GitLab, and Azure DevOps use personal access tokens to authenticate with their API:

- [GitHub personal access token](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token)
- [GitLab personal access token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html)
- [Azure DevOps personal access token](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=preview-page)

The providers require a secret in the same format, with the personal access token as the value for the token key:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: api-token
  namespace: default
data:
  token: <personal-access-tokens>
```

Bitbucket authenticates using an [app password](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/).
It requires both the username and the password when authenticating.
Therefore the token needs to be passed with the format `<username>:<app-password>`.
A token that is not in this format will cause the provider to fail.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: api-token
  namespace: default
data:
  token: <username>:<app-password>
```

Opsgenie uses an api key to authenticate [api key](https://support.atlassian.com/opsgenie/docs/api-key-management/).
The providers require a secret in the same format, with the api key as the value for the token key:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: api-token
  namespace: default
data:
  token: <api-key>
```

### Azure Event Hub

The Azure Event Hub supports two authentication methods, [JWT](https://docs.microsoft.com/en-us/azure/event-hubs/authenticate-application)
and [SAS](https://docs.microsoft.com/en-us/azure/event-hubs/authorize-access-shared-access-signature) based.

#### JWT based auth

In JWT we use 3 input values. Channel, token and address.
We perform the following translation to match we the data we need to communicate with Azure Event Hub.

* channel = Azure Event Hub namespace
* address = Azure Event Hub name
* token   = JWT

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Provider
metadata:
  name: azureeventhub
spec:
  type: azureeventhub
  channel: fluxv2
  secretRef:
    name: webhook-url
---
apiVersion: v1
data:
  address: Zmx1eHYy
  token: QS12YWxpZC1KV1QtdG9rZW4=
kind: Secret
metadata:
  name: webhook-url
  namespace: default
type: Opaque
```

Notification controller doesn't take any responsibility for the JWT token to be updated.
You need to use a secondary tool to make sure that the token in the secret is renewed.

If you want to make a easy test assuming that you have setup a Azure Enterprise application and you called it
event-hub you can follow most of the bellow commands. You will need to provide the client_secret that you got
when generating the Azure Enterprise Application.

```shell
export AZURE_CLIENT=$(az ad app list --filter "startswith(displayName,'event-hub')" --query '[].appId' |jq -r '.[0]')
export AZURE_SECRET='secret-client-secret-generated-at-creation'
export AZURE_TENANT=$(az account show -o tsv --query tenantId)

curl -X GET --data 'grant_type=client_credentials' --data "client_id=$AZURE_CLIENT" --data "client_secret=$AZURE_SECRET" --data 'resource=https://eventhubs.azure.net' -H 'Content-Type: application/x-www-form-urlencoded' https://login.microsoftonline.com/$AZURE_TENANT/oauth2/token |jq .access_token
```

Use the output you got from the curl and add it to your secret like bellow.

```shell
kubectl create secret generic webhook-url \
--from-literal=address="fluxv2" \
--from-literal=token='A-valid-JWT-token'
```

#### SAS based auth

In SAS we only use the `address` field in the secret.

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1beta1
kind: Provider
metadata:
  name: azureeventhub
spec:
  type: azureeventhub
  secretRef:
    name: webhook-url
---
apiVersion: v1
data:
  address: RW5kcG9pbnQ9c2I6Ly9mbHV4djIuc2VydmljZWJ1cy53aW5kb3dzLm5ldC87U2hhcmVkQWNjZXNzS2V5TmFtZT1Sb290TWFuYWdlU2hhcmVkQWNjZXNzS2V5O1NoYXJlZEFjY2Vzc0tleT15b3Vyc2Fza2V5Z2VuZWF0ZWRieWF6dXJlCg==
kind: Secret
metadata:
  name: webhook-url
  namespace: default
type: Opaque
```

Assuming that you have created Azure event hub and namespace you should be able to use a similar command to get your
connection string. This will give you the default Root SAS, it's NOT supposed to be used in production.

```shell
az eventhubs namespace authorization-rule keys list --resource-group <rg-name> --namespace-name <namespace-name> --name RootManageSharedAccessKey -o tsv --query primaryConnectionString
# The output should look something like this:
Endpoint=sb://fluxv2.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=yoursaskeygeneatedbyazure
```

To create the needed secret:

```shell
kubectl create secret generic webhook-url \
--from-literal=address="Endpoint=sb://fluxv2.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=yoursaskeygeneatedbyazure"
```

