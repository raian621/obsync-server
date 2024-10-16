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
e.g. `localhost:8000`, `127.0.0.1:8000`, etc.

To generate code using the OpenAPI spec, you can run

```sh
go generate
```

and the generated code will be in `api/gen.go`.
