# Contributing guidelines

## Sign the CLA

Kubernetes projects require that you sign a Contributor License Agreement (CLA) before we can accept your pull requests.

Please see https://git.k8s.io/community/CLA.md for more info

## Contributing

1. Submit an issue describing your proposed change
1. The [repo owners](OWNERS) will respond to your issue promptly.
1. Develop and test your code changes.
1. Submit a pull request.

## CI Tests

See [Travis](.travis.yml) file to check the travis tests. It is setup to run for all pull requests.
In the Pull request check the CI job `continuous-integration/travis-ci/pr` and click on `Details`.

## Changing API

This project uses and is built with [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder).
To regenerate code after changes to the [Application CRD](api/v1beta1/application_types.go), run `make generate`. Typically `make all` would take care of it. Make sure you add enough [tests](api/v1beta1/application_types_test.go). Update the [example](docs/examples/example.yaml)
