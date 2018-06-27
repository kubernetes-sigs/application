# Contributing guidelines

## Sign the CLA

Kubernetes projects require that you sign a Contributor License Agreement (CLA) before we can accept your pull requests.  Please see https://git.k8s.io/community/CLA.md for more info

## Contributing A Patch

1. Submit an issue describing your proposed change to the repo in question.
1. The [repo owners](OWNERS) will respond to your issue promptly.
1. If your proposed change is accepted, and you haven't already done so, sign a Contributor License Agreement (see details above).
1. Fork the desired repo, develop and test your code changes.
1. Submit a pull request.

### Code Generation

This project uses and is built with [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder). Kubebuilder does generates some of the code. Prior to submitting a pull request please check that you are not altering generated code and check if you need to regenerated code due to your change.

1. Make changes to the [Application CRD](pkg/apis/app/v1alpha1/application_types.go).
1. Add [tests](pkg/apis/app/v1alpha1/application_types_test.go).
1. Regenerate the generated code using `kubebuilder generate`.
1. Update the [example](docs/example.yaml)
