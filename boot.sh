#!/bin/bash

set -euo pipefail

export GOPRIVATE="github.com/adaricorp/ruckus-sz-proto"

go mod tidy
