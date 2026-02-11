#!/bin/bash
# Demonstrates creating and running a self-extracting binary.
#
# This script:
#   1. Creates sample payload files
#   2. Builds the selfextract binary
#   3. Appends payload files to the binary using gozip
#   4. Runs the binary to extract the embedded files

set -euo pipefail
cd "$(dirname "$0")"

echo "=== Step 1: Create sample payload files ==="
mkdir -p payload
echo "Hello from the self-extracting binary!" > payload/hello.txt
echo '{"app": "selfextract-demo", "version": "1.0"}' > payload/config.json
mkdir -p payload/templates
echo "<h1>Welcome</h1>" > payload/templates/index.html
echo "Payload files created."

echo ""
echo "=== Step 2: Build the selfextract binary ==="
go build -o selfextract .
echo "Binary built: ./selfextract"

echo ""
echo "=== Step 3: Build the gozip CLI tool ==="
go build -o gozip-cli ../../cmd/gozip/
echo "gozip CLI built: ./gozip-cli"

echo ""
echo "=== Step 4: Append payload to binary ==="
./gozip-cli -c selfextract payload/
echo "Payload appended to binary."

echo ""
echo "=== Step 5: List embedded files ==="
./selfextract -l

echo ""
echo "=== Step 6: Extract embedded files ==="
./selfextract -d extracted
echo ""
echo "Extracted contents:"
find extracted -type f | sort | while read -r f; do
    echo "  $f: $(cat "$f")"
done

echo ""
echo "=== Cleanup ==="
rm -rf payload extracted selfextract gozip-cli
echo "Done! All temporary files removed."
