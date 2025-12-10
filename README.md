<!--
SPDX-FileCopyrightText: Copyright 2025 SAP SE or an SAP affiliate company and cobaltcore-dev contributors

SPDX-License-Identifier: Apache-2.0
-->
# labels-injector [![REUSE status](https://api.reuse.software/badge/github.com/cobaltcore-dev/labels-injector)](https://api.reuse.software/info/github.com/cobaltcore-dev/labels-injector) [![Checks](https://github.com/cobaltcore-dev/labels-injector/actions/workflows/checks.yaml/badge.svg)](https://github.com/cobaltcore-dev/labels-injector/actions/workflows/checks.yaml) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Description

KVM labels-injector is a lightweight Kubernetes controller that reconciles Pod resources and injects predefined labels taken from the Node each Pod is scheduled on.

This repository contains the controller source, Kubernetes manifests (kustomize), and a Helm chart.

## Prerequisites

- Go 1.23+ (for building from source)
- Docker (if you want to build and push container images)
- kubectl
- Access to a Kubernetes cluster (v1.11.3+ should work; this project uses controller-runtime tooling)
- make (GNU make recommended)

## Quick developer checklist

1. Build the controller binary:

```sh
    make build-all
```

   The built binary will be placed at `build/manager`.

2. Install the binary locally (to /usr/local/bin by default):

```sh
    sudo make install
```

3. Generate Kubernetes manifests (deepcopy, CRDs, webhooks):

```sh
   make generate
```

   This will run `controller-gen` and place generated CRD artifacts under `config/crd`.

4. Produce Helm chart from kustomize (optional):

```sh
   make helm-charts
```

   This uses `helmify` to convert the kustomize output into a Helm chart under `charts/labels-injector`.

## Important make targets

- `make build-all` — build the manager binary (output: `build/manager`).
- `make install` — install `build/manager` to `$(PREFIX)/bin` (defaults to `/usr/local/bin` on macOS).
- `make generate` — generate CRDs and deepcopy code using `controller-gen`.
- `make helm-charts` — generate Helm chart from kustomize output with `helmify`.
- `make check` — run unit tests, coverage and static checks (requires some tools).
- `make tidy-deps` — run `go mod tidy` and `go mod verify`.
- `make goimports` — run `goimports` over the repo.
- `make clean` — remove build artifacts.

Some make targets automatically install required tools if missing (see the top of the Makefile). In CI you may prefer installing tools via your package manager rather than letting the Makefile install them.

## Building and publishing a container image

1. Build the binary locally:

```sh
   make build-all
```

2. Build a container image (example):

```sh
   docker build -t <registry>/labels-injector:<tag> .
```

3. Push the image:

```sh
   docker push <registry>/labels-injector:<tag>
```

4. Update the deployment manifest or Helm values to use the pushed image and deploy to your cluster.

Tips: you can use `kind load docker-image` for local testing with kind clusters instead of pushing to a remote registry.

## Running on Kubernetes

There are two supported deployment workflows: kustomize-based manifests (recommended for dev/test) and the Helm chart.

Kustomize (manifests)

1. Generate code and CRDs (if you changed the API code):

```sh
   make generate
```
2. Apply the manifests (example using the `config/default` overlay):

```sh
   kubectl apply -k config/default
```

3. If you changed the image, update the deployment (for example):

```sh
   kubectl -n <namespace> set image deployment/labels-injector manager=<registry>/labels-injector:<tag>
```

Helm

1. Generate the Helm chart (optional):

```sh
   make helm-charts
```

2. Install using Helm:

```sh
   helm install labels-injector charts/labels-injector --namespace <namespace> --create-namespace
```
3. To upgrade after building a new image or chart:

```sh
   helm upgrade labels-injector charts/labels-injector --namespace <namespace>
```

Notes on configuration

- The `config/` directory contains kustomize overlays used for generating manifests. The default overlay (`config/default`) is a good starting point for a simple install.
- RBAC, TLS/webhook configuration, metrics Service and ServiceAccount manifests are included in `config/` and in the Helm chart templates.

## Local testing

You can test the controller locally against a Kubernetes API using `envtest` (used by controller-runtime) — the project Makefile knows how to install `setup-envtest` and run tests that use it. To run the unit tests and generate a coverage report:

```sh
make build/cover.out
make build/cover.html
```

## Troubleshooting

- `controller-gen` not found: run `make install-controller-gen` or install it manually (`go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest`).
- `helmify` missing for `make helm-charts`: install it (`go install github.com/arttor/helmify/cmd/helmify@latest`) or install Helm and use the included chart directly.
- Image does not update in cluster: make sure you updated the deployment image or the Helm values and redeployed (see commands above). If using immutable image tags, bump the tag.

## Contributing

Contributions are welcome. Please follow the repository conventions and run `make goimports` and `make run-golangci-lint` before opening a PR.

## Support, Feedback, Contributing

This project is open to feature requests/suggestions, bug reports etc. via [GitHub issues](https://github.com/cobaltcore-dev/labels-injector/issues). Contribution and feedback are encouraged and always welcome. For more information about how to contribute, the project structure, as well as additional contribution information, see our [Contribution Guidelines](CONTRIBUTING.md).

## Security / Disclosure
If you find any bug that may be a security problem, please follow our instructions at [in our security policy](https://github.com/cobaltcore-dev/labels-injector/security/policy) on how to report it. Please do not create GitHub issues for security-related doubts or problems.

## Code of Conduct

We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone. By participating in this project, you agree to abide by its [Code of Conduct](https://github.com/SAP/.github/blob/main/CODE_OF_CONDUCT.md) at all times.

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Please see our [LICENSE](LICENSE) for copyright and license information.
Detailed information including third-party components and their licensing/copyright information is available [via the REUSE tool](https://api.reuse.software/info/github.com/cobaltcore-dev/kvm-node-agent).

