#!/bin/bash

# Interactive diagnostic menu with fzf
# Usage: ./menu.sh [CLUSTER_NAME]
# Requires: fzf (brew install fzf or asdf install fzf)

CLUSTER="${1:-spoke1}"

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MAKEFILE_DIR="$(dirname "$SCRIPT_DIR")"
TEST_MK_FILE="$MAKEFILE_DIR/test.mk"

# Parse test.mk to extract menu items dynamically
parse_menu_items() {
    local section=""
    local desc=""

    while IFS= read -r line; do
        # Track sections
        case "$line" in
            *"Hub Cluster Checks"*) section="[Hub] " ;;
            *"Spoke Cluster Checks"*) section="[Spoke] " ;;
        esac

        # Capture description comments
        if [[ "$line" == "# Verifica "* ]] || [[ "$line" == "# Lista "* ]]; then
            desc="${line#\# }"
        # Match check-* targets
        elif [[ "$line" == check-*: ]] && [[ -n "$desc" ]]; then
            local target="${line%%:*}"
            if [[ "$target" != "check" ]]; then
                echo "${target}|${section}${desc}"
            fi
            desc=""
        # Reset desc on non-comment lines
        elif [[ "$line" != "#"* ]]; then
            desc=""
        fi
    done < "$TEST_MK_FILE"

    echo "exit|Sair"
}

# Load menu items from test.mk
MENU_ITEMS=()
while IFS= read -r line; do
    MENU_ITEMS+=("$line")
done < <(parse_menu_items)

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check fzf
if ! command -v fzf >/dev/null 2>&1; then
    echo -e "${YELLOW}fzf is required. Install with:${NC}"
    echo "  brew install fzf"
    echo "  # or"
    echo "  asdf install fzf"
    exit 1
fi

echo -e "${GREEN}OCM Addon Diagnostic Menu${NC}"
echo "Cluster: $CLUSTER"
echo ""

while true; do
    # Extract labels for fzf
    labels=()
    for item in "${MENU_ITEMS[@]}"; do
        labels+=("${item#*|}")
    done

    # Run fzf
    selected=$(printf '%s\n' "${labels[@]}" | fzf --height=20 --reverse --no-info \
        --prompt="Select check: " \
        --header="Use arrows to navigate, Enter to select, Esc to exit")

    if [ -z "$selected" ]; then
        echo -e "${GREEN}Exiting...${NC}"
        exit 0
    fi

    # Find target for selected label
    target=""
    for item in "${MENU_ITEMS[@]}"; do
        if [ "${item#*|}" == "$selected" ]; then
            target="${item%%|*}"
            break
        fi
    done

    if [ "$target" == "exit" ]; then
        echo -e "${GREEN}Exiting...${NC}"
        exit 0
    fi

    # Run the make target
    clear
    echo -e "${GREEN}Running: $target${NC}"
    echo "Cluster: $CLUSTER"
    echo "---"
    make -f "$MAKEFILE_DIR/test.mk" "$target" CLUSTER="$CLUSTER"

    echo ""
    echo -e "Press any key to return to menu..."
    read -rsn1
done
