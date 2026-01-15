#!/bin/bash
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
echo "✅ Swagger docs generados en ./docs"
