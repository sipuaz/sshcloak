#!/usr/bin/env zsh
# sshcloak zsh completion script
#
# Usage: source this file in your ~/.zshrc or add to fpath
#
# Example 1 (direct source):
#   source ~/path/to/sshcloak/scripts/completion/zsh-completion.sh
#
# Example 2 (via fpath, recommended):
#   fpath=(~/path/to/sshcloak/scripts/completion/zsh-completion.sh $fpath)
#   autoload -U compinit && compinit
#
# The completion function queries the sshcloak binary for available hosts and
# flags. This avoids caching stale data and ensures completion stays in sync
# with your managed host list.

# Ensure we're in zsh
[[ -n "${ZSH_VERSION}" ]] || return 0

# Check zsh version; zsh completion requires 5.0+
if [[ ${ZSH_VERSION%.*} -lt 5 ]]; then
	if [[ -n "${SSHCLOAK_DEBUG_COMPLETION}" ]]; then
		echo "sshcloak completion: zsh 5.0+ required (found ${ZSH_VERSION})" >&2
	fi
	return 0
fi

# Main completion function
_sshcloak() {
	local context state line
	local -a arguments subcommands

	# Get the sshcloak binary path
	local sshcloak_bin
	if ! sshcloak_bin=$(command -v sshcloak); then
		if [[ -n "${SSHCLOAK_DEBUG_COMPLETION}" ]]; then
			echo "sshcloak completion: sshcloak not found in \$PATH" >&2
		fi
		return 1
	fi

	# Parse current command line state
	(( CURRENT == 1 )) && state="command" || state="args"

	# First argument (main subcommand)
	if (( CURRENT == 2 )); then
		local -a cmds
		cmds=(
			'completion:Print completion functions for bash and zsh'
			'connect:Connect to an SSH host via sshpass'
			'host:Add, list, inspect, edit, and remove SSH host entries'
			'init:Initialize sshcloak SSH config integration'
			'password:Manage passwords in the encrypted vault'
			'session:Manage the session cache (unlock, lock, status)'
			'session-agent:Start the session cache agent'
			'vault:Manage the encrypted secret vault'
			'version:Display sshcloak version'
		)
		_describe 'command' cmds
		return 0
	fi

	# Second level: handle subcommands
	local cmd="${words[2]}"

	case "$cmd" in
		connect)
			local hosts
			hosts=$("$sshcloak_bin" completion list-hosts 2>/dev/null | sed -n 's/{"host":"\([^"]*\)".*/\1/p' 2>/dev/null)
			if [[ -n "$hosts" ]]; then
				local -a host_list
				host_list=(${(f)hosts})
				_describe 'managed host' host_list
			fi
			;;
		host)
			if (( CURRENT == 3 )); then
				local -a subcmds
				subcmds=(
					'add:Add a new managed SSH host'
					'edit:Edit a managed SSH host'
					'get:Show details of a managed SSH host'
					'list:List all managed SSH hosts'
					'remove:Remove a managed SSH host'
					'tag:Manage host tags stored in the metadata sidecar'
				)
				_describe 'host command' subcmds
			else
				local flags
				flags=$("$sshcloak_bin" completion flags "host" 2>/dev/null)
				if [[ -n "$flags" ]]; then
					local -a flag_list
					flag_list=(${(f)flags})
					_describe 'flag' flag_list
				fi
			fi
			;;
		vault)
			if (( CURRENT == 3 )); then
				local -a subcmds
				subcmds=(
					'create-key:Create or rotate the vault encryption key'
					'init:Initialize the encrypted vault'
					'lock:Explicitly lock the vault'
					'rotate:Rotate the vault passphrase'
				)
				_describe 'vault command' subcmds
			else
				local flags
				flags=$("$sshcloak_bin" completion flags "vault" 2>/dev/null)
				if [[ -n "$flags" ]]; then
					local -a flag_list
					flag_list=(${(f)flags})
					_describe 'flag' flag_list
				fi
			fi
			;;
		password)
			if (( CURRENT == 3 )); then
				local -a subcmds
				subcmds=(
					'add:Add a password for a managed host'
					'delete:Delete a password for a managed host'
					'get:Retrieve a password for a managed host'
					'list:List all managed hosts with passwords'
				)
				_describe 'password command' subcmds
			else
				local flags
				flags=$("$sshcloak_bin" completion flags "password" 2>/dev/null)
				if [[ -n "$flags" ]]; then
					local -a flag_list
					flag_list=(${(f)flags})
					_describe 'flag' flag_list
				fi
			fi
			;;
		*)
			local flags
			flags=$("$sshcloak_bin" completion flags "$cmd" 2>/dev/null)
			if [[ -n "$flags" ]]; then
				local -a flag_list
				flag_list=(${(f)flags})
				_describe 'flag' flag_list
			fi
			;;
	esac
}

# Register the completion function
compdef _sshcloak sshcloak
