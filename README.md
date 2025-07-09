# nevr-common

the runtime framework for the NEVR service.

This codebase defines the runtime API and protocol interface used by [NEVR](https://github.com/echotools/nevr-service)

It is tightly integrated with components of [Nakama](https://github.com/heroiclabs/nakama), and is structured similarly to the `heroiclabs/nakama-common` repository.

The code is broken up into packages for different parts:

* `api` - The request/response messages used with the GPRC and in some of the real-time API.
* `rtapi`: The runtime API definitions, including the frame structure and connectivity statistics.
* `gameapi`: The game-engine's HTTP API session data and user bones structures.

## Usage

Protocol Buffer files have already been generated and are included in the repository. To use them in your project:

* **Go**: Import the generated Go packages directly in your project. Example:

    ```go
    import "github.com/echotools/nevr-common/v3/api"
    ```

* **Python**: Use the generated `.py` files in your Python project. Example:

    ```python
    from api import api_pb2
    ```

* **CSharp**: Reference the generated `.cs` files in your C# project. Example:

    ```csharp
    using Api;
    ```


No additional code generation is required unless you modify the `.proto` files.

## Generating Protocol Buffer Sources

The codebase uses Protocol Buffers. The protoc toolchain is used to generate source files which are committed to the repository to simplify builds for contributors.

To build the codebase and generate all sources use these steps.

1. Install the Go toolchain and protoc toolchain.

2. Install the protoc-gen-go plugin to generate Go code.

   ```shell
   go install "google.golang.org/protobuf/cmd/protoc-gen-go"
   ```

### Method A: Generate Using Make

To generate all source files with Make, run:

```shell
make generate
```

This command executes the required protoc commands and plugins as specified in the [Makefile](./Makefile).

To generate all source files using Make, run:

```shell
make generate
```

This command will invoke the necessary protoc commands and plugins as defined in the [Makefile](./Makefile).

## Method B: Generate Go stubs using the Go generate command

To generate all Go source files, run:

```shell
env PATH="$HOME/go/bin:$PATH" go generate -x ./...
```

These steps have been tested with the Go 1.24 toolchain. Earlier Go toolchain versions may work though YMMV.
