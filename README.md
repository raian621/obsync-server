[![codecov](https://codecov.io/gh/raian621/obsync-server/graph/badge.svg?token=VNEFZMUP9R)](https://codecov.io/gh/raian621/obsync-server)

# Obsync Plugin Server

File sync server for the Obsync plugin for Obsidian.

## Development

If you want to view the documentation page for the project's OpenAPI spec,
first download the Redoc JavaScript bundle locally by using the 
`scripts/download_redoc_bundle.sh` script:

```sh
./scripts/download_redoc_bundle.sh
```

Then, to start the server, run:

```sh
go run .
```

If you downloaded the Redoc JavaScript bundle locally, you should be able to
view the Redoc documentation page for the project's OpenAPI spec at
`<hostname>/api/v1/docs` (replace `<hostname>` with the hostname of your server,
e.g. `localhost:8000`, `127.0.0.1:8000`, etc.)

To generate code using the OpenAPI spec, you can run

```sh
go generate
```

and the generated code will be in `api/gen.go`.

## Configuration

The application can be configured via a `config.yaml` file in the project's base directory.

```yaml
# example config.yaml

type: FileSystem
root: /tmp/obsync-dev
host: localhost
port: 8000
```

### Options

- **`type`**: The type of the server's file store. Currently, there's only `FileSystem`, which uses the host machine's file system to store files. I'd like to add S3 or maybe Google Drive eventually.
- **`root`**: The root of the server's file store. When the server's file store is a `FileSystem` type, this will be the base directory where all synced files will be stored. For other future file stores, it might be an S3 bucket name or a folder in a Google Drive.
- **`host`**: The hostname that the server should listen on.
- **`port`**: The port that the server should listen on.