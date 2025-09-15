# nevr-common

the runtime framework for the NEVR service.

This codebase defines the runtime API and protocol interface used by [NEVR](https://github.com/echotools/nevr-service)

It is tightly integrated with components of [Nakama](https://github.com/heroiclabs/nakama), and is structured similarly to the `heroiclabs/nakama-common` repository.

The code is broken up into packages for different parts:

* `proto` - protobuf definitions for GRPC, HTTP, and Realtime protocols.
* `gen`   - generated source code for various languages.
* `common`: Shared utilities and types used across the codebase.


## Usage

Protocol Buffer files have already been generated and are included in the repository. To use them in your project:

* **Go**: Import the generated Go packages directly in your project. Example:

    ```go
    import "github.com/echotools/nevr-common/rtapi"
    ```

* **Python**: Use the generated `.py` files in your Python project. Example:

    ```python
    from api import api_pb
    ```

* **CSharp**: Reference the generated `.cs` files in your C# project. Example:

    ```csharp
    using Api;
    ```


No additional code generation is required unless you modify the `.proto` files.

## Generating Protocol Buffer Sources

   ```shell
   buf generate
   ```

### Method A: `build.sh`

To generate all source files with Make, run:

```shell
./build.sh
```

## Method B: Generate Go stubs using the Go generate command

To generate all Go source files, run:

```shell
env PATH="$HOME/go/bin:$PATH" go generate -x ./...
```

These steps have been tested with the Go 1.24 toolchain. Earlier Go toolchain versions may work though YMMV.
