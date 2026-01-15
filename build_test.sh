#!/bin/bash
set -e

echo "=== Building son-et transpiler ==="
go build -o son-et ./cmd/son-et
echo "✅ Transpiler built successfully"

echo ""
echo "=== Running compiler tests ==="
go test ./pkg/compiler/...
echo ""

echo "=== Running engine tests ==="
go test ./pkg/engine/...
echo ""

echo "=== Build complete ==="
ls -lh son-et
