#!/usr/bin/env bash
# sshcloak completion installation helper
#
# This script installs sshcloak completion scripts to system-wide completion
# directories (if write-accessible) or to user directories.
#
# Usage:
#   ./scripts/install-completion.sh [--bash] [--zsh] [--user]
#
# Flags:
#   --bash             Install bash completion only
#   --zsh              Install zsh completion only
#   --user             Install to user directories (~/.local/share, ~/.zsh) instead of system-wide
#   -h, --help         Print this help message

set -e

# Determine the script's directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SSHCLOAK_COMPLETION_DIR="${SCRIPT_DIR}/completion"

# Defaults
INSTALL_BASH=1
INSTALL_ZSH=1
USER_ONLY=0

# Parse arguments
while [[ $# -gt 0 ]]; do
	case "$1" in
		--bash)
			INSTALL_BASH=1
			INSTALL_ZSH=0
			shift
			;;
		--zsh)
			INSTALL_ZSH=1
			INSTALL_BASH=0
			shift
			;;
		--user)
			USER_ONLY=1
			shift
			;;
		-h | --help)
			grep '^#' "$0" | sed 's/^#[[:space:]]\{0,1\}//'
			exit 0
			;;
		*)
			echo "Unknown option: $1" >&2
			exit 1
			;;
	esac
done

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Bash completion installation
if [[ $INSTALL_BASH -eq 1 ]]; then
	echo "Installing bash completion..."

	if [[ $USER_ONLY -eq 1 ]]; then
		# User-level bash completion directory
		BASH_DIR="${HOME}/.local/share/bash-completion/completions"
		mkdir -p "$BASH_DIR"
		cp "${SSHCLOAK_COMPLETION_DIR}/bash-completion.sh" "${BASH_DIR}/sshcloak"
		echo -e "${GREEN}✓${NC} Bash completion installed to $BASH_DIR/sshcloak"
		echo "  Add to ~/.bashrc: source ${BASH_DIR}/sshcloak"
	else
		# Try system-wide bash completion directories
		if [[ -w /usr/share/bash-completion/completions ]]; then
			cp "${SSHCLOAK_COMPLETION_DIR}/bash-completion.sh" /usr/share/bash-completion/completions/sshcloak
			echo -e "${GREEN}✓${NC} Bash completion installed to /usr/share/bash-completion/completions/sshcloak"
		elif [[ -w /etc/bash_completion.d ]]; then
			cp "${SSHCLOAK_COMPLETION_DIR}/bash-completion.sh" /etc/bash_completion.d/sshcloak
			echo -e "${GREEN}✓${NC} Bash completion installed to /etc/bash_completion.d/sshcloak"
		else
			# Fall back to user directory
			BASH_DIR="${HOME}/.local/share/bash-completion/completions"
			mkdir -p "$BASH_DIR"
			cp "${SSHCLOAK_COMPLETION_DIR}/bash-completion.sh" "${BASH_DIR}/sshcloak"
			echo -e "${YELLOW}⚠${NC} No write access to system bash-completion directories."
			echo "  Installed to user directory: $BASH_DIR/sshcloak"
			echo "  Add to ~/.bashrc: source ${BASH_DIR}/sshcloak"
		fi
	fi
fi

# Zsh completion installation
if [[ $INSTALL_ZSH -eq 1 ]]; then
	echo "Installing zsh completion..."

	if [[ $USER_ONLY -eq 1 ]]; then
		# User-level zsh completion directory
		ZSH_DIR="${HOME}/.zsh/completions"
		mkdir -p "$ZSH_DIR"
		cp "${SSHCLOAK_COMPLETION_DIR}/zsh-completion.sh" "${ZSH_DIR}/_sshcloak"
		echo -e "${GREEN}✓${NC} Zsh completion installed to $ZSH_DIR/_sshcloak"
		echo "  Add to ~/.zshrc: fpath=(${ZSH_DIR} \$fpath); autoload -U compinit && compinit"
	else
		# Try system-wide zsh completion directories
		if [[ -w /usr/share/zsh/site-functions ]]; then
			cp "${SSHCLOAK_COMPLETION_DIR}/zsh-completion.sh" /usr/share/zsh/site-functions/_sshcloak
			echo -e "${GREEN}✓${NC} Zsh completion installed to /usr/share/zsh/site-functions/_sshcloak"
		elif [[ -w /usr/local/share/zsh/site-functions ]]; then
			cp "${SSHCLOAK_COMPLETION_DIR}/zsh-completion.sh" /usr/local/share/zsh/site-functions/_sshcloak
			echo -e "${GREEN}✓${NC} Zsh completion installed to /usr/local/share/zsh/site-functions/_sshcloak"
		else
			# Fall back to user directory
			ZSH_DIR="${HOME}/.zsh/completions"
			mkdir -p "$ZSH_DIR"
			cp "${SSHCLOAK_COMPLETION_DIR}/zsh-completion.sh" "${ZSH_DIR}/_sshcloak"
			echo -e "${YELLOW}⚠${NC} No write access to system zsh completion directories."
			echo "  Installed to user directory: $ZSH_DIR/_sshcloak"
			echo "  Add to ~/.zshrc: fpath=(${ZSH_DIR} \$fpath); autoload -U compinit && compinit"
		fi
	fi
fi

echo ""
echo -e "${GREEN}Installation complete!${NC}"
echo "Reload your shell or source your rc file to activate completion."
