# FIAS Parser

CLI utility for streaming GAR XML data to JSONL.

## Build Docker image
```
docker build -t fias-parser .
```

## Run from Docker
Mount your XML directory and point the CLI to it. Schemas are already baked into the image at `/gar_schemas`.
```
docker run --rm \
  -v "$(pwd)/test:/data" \
  fias-parser \
  --xml /data \
  --schema-dir /gar_schemas
```

The command above streams every XML file in the host `test` folder to stdout. Redirect stdout to capture JSONL records, e.g.:
```
docker run --rm \
  -v "$(pwd)/test:/data" \
  fias-parser \
  --xml /data \
  --schema-dir /gar_schemas > output.jsonl
```

If your XML uses a specific child element under the root, add `--element <NAME>`.
