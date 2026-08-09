#!/bin/bash

set -e

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)

VERSION=$(node -p "require('$ROOT_DIR/package.json').version")

echo "Building version: $VERSION"

RELEASE_DIR="$ROOT_DIR/releases/server"
SERVER_DIR="$ROOT_DIR/apps/backend/cmd/server"



echo "Cleaning server release"

rm -rf "$RELEASE_DIR"

mkdir -p "$RELEASE_DIR"


echo "Installing dependencies"

cd "$ROOT_DIR"

yarn install


echo "Building contracts"

cd "$ROOT_DIR/packages/contracts"

yarn build


echo "Building components"

cd "$ROOT_DIR/packages/ui-components"

yarn build


echo "Building SDK"

cd "$ROOT_DIR/packages/sdk"

yarn build


echo "Building frontend"

cd "$ROOT_DIR/apps/frontend"

VERSION=$VERSION yarn build


echo "Building backend"

cd "$ROOT_DIR/apps/backend"

go build \
-ldflags "-X lwn-simulator-backend/version.AppVersion=$VERSION" \
-o "$RELEASE_DIR/lwn-server" \
cmd/server/main.go


echo "Server release created"

ls -lah "$RELEASE_DIR"