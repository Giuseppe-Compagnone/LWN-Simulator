#!/bin/bash

set -e

echo "Generating contracts..."

cd packages/contracts

yarn generate

echo "Contracts generated"