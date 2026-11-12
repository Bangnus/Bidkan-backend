#!/usr/bin/env bash
# ==============================================================================
# BidKan Platform — DigitalOcean Droplet Automated Deployment Script
# OS Target: Ubuntu 22.04 / 24.04 LTS
# ==============================================================================

set -euo pipefail

echo "======================================================"
echo "🚲 Starting BidKan Platform Deployment on Droplet..."
echo "======================================================"

# 1. Check Root Privileges
if [ "$EUID" -ne 0 ]; then
  echo "❌ Please run this script with sudo or as root: sudo bash deploy-droplet.sh"
  exit 1
fi

# 2. Setup 2GB Swap Memory (Prevents OOM during Next.js compile on 2GB RAM droplets)
TOTAL_RAM_KB=$(grep MemTotal /proc/meminfo | awk '{print $2}')
if [ "$TOTAL_RAM_KB" -lt 4000000 ] && [ ! -f /swapfile ]; then
  echo "🧠 Creating 2GB Swap file for build stability..."
  fallocate -l 2G /swapfile
  chmod 600 /swapfile
  mkswap /swapfile
  swapon /swapfile
  echo '/swapfile none swap sw 0 0' >> /etc/fstab
  echo "✅ 2GB Swap memory activated."
fi

# 3. Install Docker & Docker Compose if missing
if ! command -v docker &> /dev/null; then
  echo "🐳 Installing Docker Engine..."
  apt-get update -y
  apt-get install -y ca-certificates curl gnupg lsb-release
  mkdir -p /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
  apt-get update -y
  apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  systemctl enable docker
  systemctl start docker
  echo "✅ Docker installed successfully."
fi

# 4. Configure UFW Firewall
if command -v ufw &> /dev/null; then
  echo "🛡️ Configuring Firewall rules (UFW)..."
  ufw allow 22/tcp    # SSH
  ufw allow 80/tcp    # HTTP (Let's Encrypt & Web)
  ufw allow 443/tcp   # HTTPS
  ufw allow 1883/tcp  # Mosquitto MQTT for IoT ESP32 4G
  ufw --force enable
  echo "✅ UFW Firewall configured."
fi

# 5. Check Environment File
if [ ! -f .env.production ]; then
  if [ -f .env.production.example ]; then
    echo "⚠️ .env.production not found! Copying from .env.production.example..."
    cp .env.production.example .env.production
    echo "❗ Please edit .env.production with your real domain and secrets before launching!"
    exit 1
  else
    echo "❌ Error: Neither .env.production nor .env.production.example found!"
    exit 1
  fi
fi

# 6. Build and Launch Containers
echo "🚀 Building and starting Docker containers..."
docker compose --env-file .env.production -f docker-compose.prod.yml up -d --build

echo ""
echo "======================================================"
echo "🎉 BidKan Platform Launched Successfully!"
echo "======================================================"
docker compose -f docker-compose.prod.yml ps

echo ""
echo "📌 Next Steps:"
echo "1. Run Certbot to generate your free SSL certificates:"
echo "   docker compose -f docker-compose.prod.yml run --rm certbot certonly --webroot -w /var/www/certbot -d yourdomain.com -d api.yourdomain.com -d admin.yourdomain.com -d app.yourdomain.com"
echo "2. Reload Nginx to activate SSL:"
echo "   docker compose -f docker-compose.prod.yml exec nginx nginx -s reload"
echo "3. Monitor live logs:"
echo "   docker compose -f docker-compose.prod.yml logs -f"
