# helm-plugin-ocigetter

Prototype Helm OCI getter WASM/WASI plugin based on Extism WASM plugin framework

The plugin uses a modifed version of Helm's existing `registry.Client` (specifically the ORAS v2 version from https://github.com/helm/helm/pull/13382). As a proof-of-concept in getting a WASM based plugin build for Helm (targeting existing functionality)

## Build / test

The plugin is built with go (standard/"big" go)

```
make build
make test
```

## Helm integration

_(Work in progress, will post PR when ready)_

## Notes

Once go v.14 is released (expected: Feb), go will support a `//go:wasmexport name` directive, allowing WASM modules to export reactor style functions:
- <https://tip.golang.org/doc/go1.24#wasm>
- <https://github.com/golang/go/issues/65199>