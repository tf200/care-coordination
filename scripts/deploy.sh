#!/bin/bash



# 2. Connect to server and update
echo "Connecting to server to pull and rebuild..."
ssh root@maicare.online << 'EOF'
    cd care-coordination/
    git pull origin main
    make docker-rebuild
    exit
EOF

echo "✅ Deployment finished successfully!"