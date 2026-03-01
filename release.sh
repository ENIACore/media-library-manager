#!/usr/bin/env bash

VERSION="v0.9.9"
NOTES="Initial release"

gh release create ${VERSION} \
  dist/mlm-linux-amd64 \
  dist/mlm-linux-arm64 \
  dist/mlm-darwin-amd64 \
  dist/mlm-darwin-arm64 \
  dist/mlm-windows-amd64.exe \
  --title ${VERSION} \
  --notes ${NOTES}
