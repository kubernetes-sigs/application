# Releasing

[semver]: https://semver.org
[`VERSION`]: ../VERSION
[`VERSION-DEV`]: ../VERSION-DEV

This document describes how to perform a [semver] release.

## Release artifacts
1. Branch `release-vMajor.Minor` for each `Major.Minor.*` version. Example `release-v0.8`
2. Tag `vMajor.Minor.Patch` for each `Major.Minor.Patch` version. Example `v0.8.1`
3. All in one deployment yaml for the Application controller
4. Container image for the Application controller

## Release roles

Only Repo Owners can create branches in the [Application Repo](https://github.com/kubernetes-sigs/application). Developers who fork the repo can create releases in their repo as well.

#### Version files
For official releases [`VERSION`] file is used.
For developers [`VERSION-DEV`] file is used.

The default file used is [`VERSION-DEV`].
To use [`VERSION`], set the `VERSION_FILE` env variable.

Developers should edit the [`VERSION-DEV`] file to set their choice of container registry and version.

## Release procedure

#### Create release branch

Release are always cut from the `master` branch `HEAD`. 
Ensure that all necessary fixes are merged, documentation updated and most importantly the [`VERSION`] file is updated.
The steps to create a release branch are:
```bash

# Repo owners use this for official releases:
VERSION_FILE=VERSION make release-branch

# Developers use this for their fork
make release-branch
```

#### Create release tag

Patch releases are created from the patch branch. 
Ensure that all necessary fixes are merged, documentation updated and most importantly the `patch` version is updated in the [`VERSION`] file.
The steps to create a release tag are:
```bash

# Repo owners use this for official releases:
VERSION_FILE=VERSION make release-tag

# Developers use this for their fork
make release-tag
```

#### Deleting release tag

The steps to delete a release tag are:
```bash

# Repo owners use this for official releases:
VERSION_FILE=VERSION make delete-release-tag

# Developers use this for their fork
make delete-release-tag
```

### TODO
- Release notes
- Generate changes between releases

