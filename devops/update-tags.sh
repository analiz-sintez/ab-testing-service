#!/bin/bash

# Set working directory to repo location
cd /home/spl/ab-testing-service || exit 1

# Get current git tag or commit hash if no tag
GIT_TAG=$(git describe --tags --abbrev=0 2>/dev/null || exit 1)

# Create .env file with tag information
cat > .env << EOF
BACKEND_TAG=${GIT_TAG}
FRONTEND_TAG=${GIT_TAG}
MIGRATOR_TAG=${GIT_TAG}

BACKEND_PORT=38080
SERVICE_PORT=38081
FRONTEND_PORT=38000

POSTGRES_PORT=35432
REDIS_PORT=36379
KAFKA_PORT=39092

PROMETHEUS_PORT=39090
GRAFANA_PORT=28000
EOF

echo ">> Created .env with tag: ${GIT_TAG}"
