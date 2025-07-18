#!/bin/bash

# User Management System - Quick Start Script
# For daily development use after initial setup

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}🚀 Starting User Management System...${NC}"
echo

# Check if databases are running
if ! docker ps | grep -q user_mgmt_postgres; then
    echo -e "${YELLOW}⚙️ Starting databases...${NC}"
    make docker-up
    echo -e "${GREEN}✅ Databases started${NC}"
    echo
    
    # Wait a moment for databases to be ready
    echo -e "${BLUE}⏳ Waiting for databases to be ready...${NC}"
    sleep 8
else
    echo -e "${GREEN}✅ Databases already running${NC}"
fi

# Start the server
echo -e "${BLUE}🌐 Starting server...${NC}"
echo

# Check if port is specified
if [ -n "$1" ]; then
    echo -e "${GREEN}📡 Server will start on port $1${NC}"
    PORT=$1 make run
else
    echo -e "${GREEN}📡 Server will start on port 8080${NC}"
    make run
fi 