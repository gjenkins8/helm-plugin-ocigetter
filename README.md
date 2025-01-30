# helm-plugin-ocigetter

Prototype Helm OCI getter WASM/WASI plugin based on Extism WASM plugin framework

The plugin uses a modifed version of Helm's existing `registry.Client` (specifically the ORAS v2 version from https://github.com/helm/helm/pull/13382). As a proof-of-concept in getting a WASM based plugin build for Helm (targeting existing functionality)

## Build / test

```
make build
make test
```
