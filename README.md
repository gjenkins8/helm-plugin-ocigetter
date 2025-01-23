# helm-plugin-ocigetter

Prototype Helm OCI getter WASM/WASI plugin based on Extism WASM plugin framework

The plugin uses a modifed version of Helm's existing `registry.Client` (specifically the ORAS v2 version from https://github.com/helm/helm/pull/13382). As a proof-of-concept in getting a WASM based plugin build for Helm (targeting existing functionality)

## Build / test

_A modified version of tinygo is require, see below_

```
make build
make test
```

## Custom tinygo

Currently a modified version of the tinygo compiler with an extended std library is required to build the plugin.

Follow the tinygo build instructions: https://tinygo.org/docs/guides/build/> (_Manual LLVM build_ variant)

One cavet is that tinygo doesn't seem compatible with llvm version 19 (latest release). And needs to be built with llvm v18.
On my macbook, this looks like:

```
git clone https://github.com/gjenkins8/tinygo
git checkout stdlib_extensions
git submodule update --init --recursive

brew install llvm@18
make llvm-source llvm-build CLANG_SRC=/opt/homebrew/opt/llvm@18 LLD_SRC=/opt/homebrew/opt/llvm@18
make clean wasi-libc tinygo CLANG_SRC=/opt/homebrew/opt/llvm@18 LLD_SRC=/opt/homebrew/opt/llvm@18 |& grep -v "ld: warning: object file"
```

Then add `./build/tinygo` to your PATH:
```
export PATH="$(pwd)/build/tinygo:${PATH}"


## Extism Go PDK Plugin

See more documentation at https://github.com/extism/go-pdk and
[join us on Discord](https://extism.org/discord) for more help.

