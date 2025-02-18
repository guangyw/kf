---
title: App Manifest
description: "Reference guide for the application manifest.yml format."
---

App manifests provide a way for developers to record their App's execution environment in a declarative way.
They allow Apps to be deployed consistently and reproducibly.

## Format

Manifests are YAML files in the root directory of the App. They **must** be named `manifest.yml` or `manifest.yaml`.

Kf App manifests are allowed to have a single top-level element: `applications`.
The `applications` element can contain one or more application entries.

## Application fields

The following fields are valid for objects under `applications`:

| Field                        | Type       | Description |
| ---                          | ---        | ---         |
| `name`                       | `string`   | The name of the application. The app name should be lower-case alphanumeric characters and dashes. It must not start with a dash. |
| `path`                       | `string`   | The path to the source of the app. Defaults to the manifest's directory. |
| `buildpacks`                 | `string[]` | A list of buildpacks to apply to the app. |
| `stack`                      | `string`   | Base image to use for to use for apps created with a buildpack. |
| `docker`                     | `object`   | A docker object. See the Docker Fields section for more information. |
| `env`                        | `map`      | Key/value pairs to use as the environment variables for the app and build. |
| `services`                   | `string[]` | A list of service instance names to automatically bind to the app. |
| `disk_quota`                 | `quantity` | The amount of disk the application should get. Defaults to 1GiB. |
| `memory`                     | `quantity` | The amount of RAM to provide the app. Defaults to 1GiB. |
| `cpu` †                      | `quantity` | The amount of CPU to provide the application. Defaults to 0.1 (1/10th of a CPU). |
| `instances`                  | `int`      | The number of instances of the app to run. Defaults to 1. |
| `routes`                     | `object`   | A list of routes the app should listen on. See the Route Fields section for more. |
| `no-route`                   | `boolean`  | If set to true, the application will not be routable. |
| `random-route`               | `boolean`  | If set to true, the app will be given a random route. |
| `timeout`                    | `int`      | The number of seconds to wait for the app to become healthy. |
| `health-check-type`          | `string`   | The type of health-check to use `port`, `process`, `none`, or `http`. Default: `port` |
| `health-check-http-endpoint` | `string`   | The endpoint to target as part of the health-check. Only valid if `health-check-type` is `http`. |
| `command`                    | `string`   | The command that starts the app. If supplied, this will be passed to the container entrypoint. |
| `entrypoint` †               | `string`   | Overrides the app container's entrypoint. |
| `args` †                     | `string[]` | Overrides the arguments the app container. |
| `ports` †                    | `object`   | A list of ports to expose on the container. If supplied, the first entry in this list is used as the default port. |

† Unique to Kf


## Docker fields

The following fields are valid for `application.docker` objects:

| Field   | Type     | Description |
| ---     | ---      | ---         |
| `image` | `string` | The docker image to use. |

## Route fields

The following fields are valid for `application.routes` objects:

| Field     | Type     | Description |
| ---       | ---      | ---         |
| `route`   | `string` | A route to the app including hostname, domain, and path. |
| `appPort` | `int`    | (Optional) A custom port on the App the route will send traffic to. |

{{< note >}} If you specify an `appPort`, that port MUST also be declared in the `ports` field.{{< /note >}}

## Port fields

The following fields are valid for `application.ports` objects:

| Field      | Type     | Description |
| ---        | ---      | ---         |
| `port`     | `int`    | The port to expose on the App's container. |
| `protocol` | `string` | The protocol of the port to expose. Must be `tcp`, `http` or `http2`. Default: `tcp` |

{{< note >}}  The `protocol` field is a hint about the traffic that goes over the port.
The hint is used by the [service mesh](https://cloud.google.com/service-mesh/docs/overview) for better tracing and metrics.{{< /note >}}

{{< warning >}} Kf doesn't currently support TCP port-based routing. You must use a
[Kubernetes LoadBalancer](https://kubernetes.io/docs/tutorials/stateless-application/expose-external-ip-address/) if you want to expose a TCP port to the Internet. Ports are available on the cluster internal App address `<app-name>.<space>`.{{< /warning >}}



## Examples

### Minimal App

This is a bare-bones manifest that will build an App by auto-detecting
the buildpack based on the uploaded source, and deploy one instance of it.

``` yaml
---
applications:
- name: my-minimal-application
```

### Simple App

This is a full manifest for a more traditional Java App.

``` yaml
---
applications:
- name: account-manager
  # only upload src/ on push
  path: src
  # use the Java buildpack
  buildpacks:
  - java
  env:
    # manually configure the buildpack's Java version
    BP_JAVA_VERSION: 8
    ENVIRONMENT: PRODUCTION
  # use less disk and memory than default
  disk_quota: 512M
  memory: 512M
  # bump up the CPU
  cpu: 0.2
  instances: 3
  # make the app listen on three routes
  routes:
  - route: accounts.mycompany.com
  - route: accounts.datacenter.mycompany.internal
  - route: mycompany.com/accounts
  # set up a longer timeout and custom endpoint to validate
  # when the app comes up
  timeout: 300
  health-check-type: http
  health-check-http-endpoint: /healthz
  # attach two services by name
  services:
  - customer-database
  - web-cache
```

### Docker App

Kf can deploy Docker containers as well as manifest deployed App.
These Docker Apps MUST listen on the `PORT` environment variable.

``` yaml
---
applications:
- name: white-label-app
  # use a pre-built docker image (must listen on $PORT)
  docker:
    image: gcr.io/my-company/white-label-app:123
  env:
    # add additional environment variables
    ENVIRONMENT: PRODUCTION
  disk_quota: 1G
  memory: 1G
  cpu: 2
  instances: 1
  routes:
  - route: white-label-app.mycompany.com
```

### App with multiple ports

This App has multiple ports to expose an admin console, website, and SMTP server.

``` yaml
---
applications:
- name: b2b-server
  ports:
  - port: 8080
    protocol: http
  - port: 9090
    protocol: http
  - port: 2525
    protocol: tcp
  routes:
  - route: b2b-admin.mycompany.com
    appPort: 9090
  - route: b2b.mycompany.com
    # gets the default (first) port
```

### Health check types

Kf supports three different health check types:

1. `port` (default)
1. `http`
1. `process` (or `none`)

`port` and `http` set a [Kubernetes readiness and liveness
probe](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
that ensures the application is ready before sending traffic to it.

The `port` health check will ensure the port found at `$PORT` is being
listened to. Under the hood Kf uses a TCP probe.

The `http` health check will use the configured value in
`health-check-http-endpoint` to check the application's health. Under the hood
Kf uses an HTTP probe.

A `process` health check only checks to see if the process running on the
container is alive. It does NOT set a Kubernetes readiness or liveness probe.

## Known differences

The following are known differences between Kf manifests and CF manifests:

* Kf does not support deprecated CF manifest fields. This includes all fields at the root-level of the manifest (other than applications) and routing fields.
* Kf is missing support for the following v2 manifest fields:
  * `docker.username`
* Kf does not support auto-detecting ports for Docker containers.
