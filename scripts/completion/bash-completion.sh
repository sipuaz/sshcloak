#!/usr/bin/env bash
# sshcloak bash completion script
#
# Usage: source this file in your ~/.bashrc
#
# Example:
#   source ~/path/to/sshcloak/scripts/completion/bash-completion.sh
#
# The completion function queries the sshcloak binary for available hosts and
# flags. This avoids caching stale data and ensures completion stays in sync
# with your managed host list.

# Check bash version; bash completion requires 4.4+
if [[ ${BASH_VERSINFO[0]} -lt 4 ]] || [[ ${BASH_VERSINFO[0]} -eq 4 && ${BASH_VERSINFO[1]} -lt 4 ]]; then
	# macOS ships with bash 3.2 by default; print error and skip setup
	if [[ -n "${SSHCLOAK_DEBUG_COMPLETION}" ]]; then
		echo "sshcloak completion: bash 4.4+ required (found ${BASH_VERSINFO[0]}.${BASH_VERSINFO[1]})" >&2
	fi
	return 0 2>/dev/null || exit 0
fi

# Main completion function
_sshcloak_completion() {
	local cur cword
	cur="${COMP_WORDS[COMP_CWORD]}"
	cword=$COMP_CWORD

	local sshcloak_bin
	if ! sshcloak_bin=$(command -v sshcloak); then
		if [[ -n "${SSHCLOAK_DEBUG_COMPLETION}" ]]; then
			echo "sshcloak completion: sshcloak not found in \$PATH" >&2
		fi
		return 1
	fi

	# Top-level: sshcloak <TAB>
	if [[ $cword -eq 1 ]]; then
		local subcommands='completion connect host init password session session-agent vault version'
		COMPREPLY=($(compgen -W "$subcommands" -- "$cur"))
		return 0
	fi

	local cmd="${COMP_WORDS[1]}"

	case "$cmd" in
		connect)
			# sshcloak connect <TAB> — suggest managed host aliases
			local hosts
			hosts=$("$sshcloak_bin" completion list-hosts 2>/dev/null | sed -n 's/{"host":"\([^"]*\)".*/\1/p')
			if [[ -n "$hosts" ]]; then
				COMPREPLY=($(compgen -W "$hosts" -- "$cur"))
			fi
			;;
		host)
			if [[ $cword -eq 2 ]]; then
				COMPREPLY=($(compgen -W "add edit get list remove tag" -- "$cur"))
			elif [[ "$cur" == -* ]]; then
				local flags
				flags=$("$sshcloak_bin" completion flags "host" 2>/dev/null)
				COMPREPLY=($(compgen -W "$flags" -- "$cur"))
			fi
			;;
		vault)
			if [[ $cword -eq 2 ]]; then
				COMPREPLY=($(compgen -W "create-key init lock rotate" -- "$cur"))
			elif [[ "$cur" == -* ]]; then
				local flags
				flags=$("$sshcloak_bin" completion flags "vault" 2>/dev/null)
				COMPREPLY=($(compgen -W "$flags" -- "$cur"))
			fi
			;;
		password)
			if [[ $cword -eq 2 ]]; then
				COMPREPLY=($(compgen -W "add delete get list" -- "$cur"))
			elif [[ "$cur" == -* ]]; then
				local flags
				flags=$("$sshcloak_bin" completion flags "password" 2>/dev/null)
				COMPREPLY=($(compgen -W "$flags" -- "$cur"))
			fi
			;;
		session)
			if [[ $cword -eq 2 ]]; then
				COMPREPLY=($(compgen -W "lock status unlock" -- "$cur"))
			fi
			;;
		*)
			if [[ "$cur" == -* ]]; then
				local flags
				flags=$("$sshcloak_bin" completion flags "$cmd" 2>/dev/null)
				COMPREPLY=($(compgen -W "$flags" -- "$cur"))
			fi
			;;
	esac

	return 0
}

# Register the completion function
complete -o bashdefault -o default -F _sshcloak_completion sshcloak
