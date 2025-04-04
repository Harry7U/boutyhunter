#!/bin/bash
# run-bounty.sh

DOMAIN="$1"
TOOLS_DIR="$HOME/.bounty_tools"
OUTPUT_DIR="$HOME/bounty_output"
BINARY_PATH="$TOOLS_DIR/bountyhunter"

# Dependency check
check_deps() {
    declare -A deps=(
        ["go"]="https://golang.org/dl/"
        ["git"]="https://git-scm.com/"
        ["pip3"]="https://pip.pypa.io/"
    )
    
    for dep in "${!deps[@]}"; do
        if ! command -v $dep &> /dev/null; then
            echo "✖ Missing dependency: $dep - install from ${deps[$dep]}"
            exit 1
        fi
    done
}

# Build binary
build_tool() {
    echo "🛠  Building BountyHunter..."
    mkdir -p "$TOOLS_DIR"
    go build -o "$BINARY_PATH" bountyhunter/main.go
    chmod +x "$BINARY_PATH"
}

# Main execution
main() {
    check_deps
    
    if [ ! -f "$BINARY_PATH" ]; then
        build_tool
    fi
    
    if [ -z "$DOMAIN" ]; then
        echo "Usage: $0 <domain>"
        exit 1
    fi
    
    "$BINARY_PATH" "$DOMAIN" --parallel
    
    if [ -n "$WEBHOOK_URL" ]; then
        echo "📡 Sending notification to $WEBHOOK_URL"
        curl -s -X POST -H "Content-Type: application/json" \
            -d "{\"domain\":\"$DOMAIN\",\"status\":\"completed\"}" \
            "$WEBHOOK_URL"
    fi
}

main
