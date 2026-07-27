#!/bin/bash

set -e


ROOT_DIR=$(pwd)

RELEASE_DIR="$ROOT_DIR/releases/server"

SERVER_DIR="$ROOT_DIR/apps/backend/cmd/server"


echo "Cleaning release directory"

rm -rf "$RELEASE_DIR"
mkdir -p "$RELEASE_DIR"


echo "Installing dependencies"

yarn install

echo "Building contracts"

cd "$ROOT_DIR/packages/contracts"

yarn build


echo "Building frontend"

cd "$ROOT_DIR/apps/frontend"

yarn build


echo "Copying frontend for embedding"

rm -rf "$SERVER_DIR/web"

mkdir -p "$SERVER_DIR/web"

cp -r out/* "$SERVER_DIR/web/"


echo "Removing empty directories"

find "$SERVER_DIR/web" -type d -empty -delete


echo "Building backend"

cd "$ROOT_DIR/apps/backend"


go build \
-o "$RELEASE_DIR/lwn-server" \
cmd/server/main.go


echo "Cleaning embedded files"

rm -rf "$SERVER_DIR/web"


echo "Server release created"

ls -lah "$RELEASE_DIR"