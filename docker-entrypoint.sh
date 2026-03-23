#!/bin/sh
set -e

# Start cloudflared tunnel in background (if token is provided)
if [ -n "$CLOUDFLARE_TUNNEL_TOKEN" ]; then
    echo "Starting CloudFlare Tunnel..."
    cloudflared tunnel --no-autoupdate run --token $CLOUDFLARE_TUNNEL_TOKEN &
else
    echo "Warning: CLOUDFLARE_TUNNEL_TOKEN not set, tunnel not started"
fi

# Start the application in foreground
echo "Starting application..."
exec /app/ctc-api