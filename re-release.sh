#!/usr/bin/env bash

VERSION="v1.1-beta"
NOTES="First beta release - lacks documentation and extensive testing"

gh release upload ${VERSION} \
  dist/mlm-linux-amd64 \
  dist/mlm-linux-arm64 \
  dist/mlm-darwin-amd64 \
  dist/mlm-darwin-arm64 \
  dist/mlm-windows-amd64.exe \
  --clobber
