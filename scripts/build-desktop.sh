#!/bin/bash

set -e


ROOT_DIR=$(pwd)

RELEASE_DIR="$ROOT_DIR/releases/desktop"

# Electron assets
DESKTOP_ASSETS="$ROOT_DIR/apps/desktop/assets"

# Directory dove vive main.go con //go:embed
SERVER_DIR="$ROOT_DIR/apps/backend/cmd/server"


echo "Cleaning desktop release"

rm -rf "$RELEASE_DIR"
rm -rf "$DESKTOP_ASSETS"

mkdir -p "$RELEASE_DIR"
mkdir -p "$DESKTOP_ASSETS"


echo "Installing dependencies"

yarn install

echo "Building contracts"

cd "$ROOT_DIR/packages/contracts"

yarn build

echo "Building frontend"

cd "$ROOT_DIR/apps/frontend"

yarn build


echo "Copying frontend for Go embed"

rm -rf "$SERVER_DIR/web"

mkdir -p "$SERVER_DIR/web"

cp -r out/* "$SERVER_DIR/web/"


echo "Building backend"

cd "$ROOT_DIR/apps/backend"

go build \
-o "$DESKTOP_ASSETS/lwn-server" \
cmd/server/main.go


echo "Cleaning embedded frontend"

rm -rf "$SERVER_DIR/web"


echo "Building Electron"

cd "$ROOT_DIR/apps/desktop"

yarn build


echo "Copying release"

cp dist/*.AppImage "$RELEASE_DIR/"


echo "Desktop release created"

ls -lah "$RELEASE_DIR"