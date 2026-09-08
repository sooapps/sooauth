#!/usr/bin/env bash
# Sooauth Self-Host Installer
# One-line setup for Docker Compose self-hosting
set -euo pipefail

CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${CYAN}"
cat << "EOF"
   ___                 _   _     
  / __|___  ___  __ _ | | | |_   
  \__ \ (_) / _ \/ _` | |_| ' \  
  |___/\___/\___/\__,_|\__,_|_||_|
  Open-Source Authentication Layer
EOF
echo -e "${NC}"

echo -e "${CYAN}==>${NC} Checking prerequisites..."

# 1. Check Docker
if ! command -v docker >/dev/null 2>&1; then
  echo -e "${RED}Error:${NC} Docker is not installed. Please install Docker first: https://docs.docker.com/get-docker/"
  exit 1
fi

if ! docker info >/dev/null 2>&1; then
  echo -e "${RED}Error:${NC} Docker daemon is not running. Please start Docker and try again."
  exit 1
fi

# 2. Check Docker Compose
DOCKER_COMPOSE_CMD=""
if docker compose version >/dev/null 2>&1; then
  DOCKER_COMPOSE_CMD="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
  DOCKER_COMPOSE_CMD="docker-compose"
else
  echo -e "${RED}Error:${NC} Neither 'docker compose' nor 'docker-compose' was found."
  exit 1
fi

echo -e "${GREEN}✓${NC} Docker and Compose are ready."

# 3. Setup Working Directory
INSTALL_DIR="${SOOAUTH_DIR:-$PWD}"
if [ ! -f "$INSTALL_DIR/docker-compose.yml" ]; then
  # Running from curl/outside repo: create directory
  INSTALL_DIR="${SOOAUTH_DIR:-$PWD/sooauth}"
  echo -e "${CYAN}==>${NC} Setting up Sooauth in ${INSTALL_DIR}..."
  mkdir -p "$INSTALL_DIR"
  cd "$INSTALL_DIR"

  echo -e "${CYAN}==>${NC} Downloading docker-compose.yml..."
  curl -sSL -o docker-compose.yml https://raw.githubusercontent.com/sooapps/sooauth/main/docker-compose.yml
else
  cd "$INSTALL_DIR"
fi

# 4. Generate Random Secrets
generate_secret() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 32
  else
    head -c 32 /dev/urandom | od -A n -t x1 | tr -d ' \n'
  fi
}

# 5. Create .env if missing
if [ ! -f .env ]; then
  echo -e "${CYAN}==>${NC} Generating secure environment configuration (.env)..."
  MFA_KEY=$(generate_secret)
  PG_PASS=$(generate_secret)

  cat > .env << EOF
PORT=8080
SOOAUTH_ENV=production
APP_URL=http://localhost:8080
API_URL=http://localhost:8080
POSTGRES_USER=sooauth
POSTGRES_PASSWORD=${PG_PASS}
POSTGRES_DB=sooauth
DATABASE_URL=postgres://sooauth:${PG_PASS}@postgres:5432/sooauth?sslmode=disable
REDIS_URL=redis://redis:6379/0
AUTO_MIGRATE=true
MFA_ENCRYPTION_KEY=${MFA_KEY}
ADMIN_EMAILS=
WEBAUTHN_RP_ID=localhost
EOF
  echo -e "${GREEN}✓${NC} Generated .env with random cryptographic secrets."
else
  echo -e "${YELLOW}!${NC} Existing .env detected. Keeping current configuration."
fi

# 6. Start Services
echo -e "${CYAN}==>${NC} Starting Sooauth containers..."
$DOCKER_COMPOSE_CMD pull || true
$DOCKER_COMPOSE_CMD up -d

# 7. Final Summary
echo ""
echo -e "${GREEN}====================================================${NC}"
echo -e "${GREEN}  Sooauth is up and running!                        ${NC}"
echo -e "${GREEN}====================================================${NC}"
echo ""
echo -e "  ${CYAN}Sign-in & Sign-up:${NC}    http://localhost:8080/auth/sign-in"
echo -e "  ${CYAN}Admin Panel:${NC}          http://localhost:8080/admin/"
echo -e "  ${CYAN}OIDC Discovery:${NC}       http://localhost:8080/.well-known/openid-configuration"
echo ""
echo -e "To view logs:"
echo -e "  ${YELLOW}$DOCKER_COMPOSE_CMD logs -f sooauth${NC}"
echo ""
echo -e "To stop Sooauth:"
echo -e "  ${YELLOW}$DOCKER_COMPOSE_CMD down${NC}"
echo ""
