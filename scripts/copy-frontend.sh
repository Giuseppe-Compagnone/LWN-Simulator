#!/bin/bash

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
SERVER_DIR="$ROOT_DIR/apps/backend"

echo "$SERVER_DIR"

echo "Copying frontend for embedding"

rm -rf "$SERVER_DIR/internal/frontend/web"

mkdir -p "$SERVER_DIR/internal/frontend/web"

cp -r out/* "$SERVER_DIR/internal/frontend/web/"


echo "Removing empty directories"

find "$SERVER_DIR/internal/frontend/web" -type d -empty -delete