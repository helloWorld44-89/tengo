package main

const bashCompletion = `# bash completion for tengo
# Add to ~/.bash_completion or source from ~/.bashrc:
#   source <(tengo -completion bash)

_tengo_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="-F -R -validate -format -L -key -set -type -version -update -dry-run -stdout -backup -q -regex -confirm-all -completion"

    case "${prev}" in
        -type)
            COMPREPLY=( $(compgen -W "yaml yml json toml xml ini" -- "${cur}") )
            return 0
            ;;
        -completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- "${cur}") )
            return 0
            ;;
        -F|-R|-key|-set|-L)
            return 0
            ;;
    esac

    if [[ "${cur}" == -* ]]; then
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
    else
        COMPREPLY=( $(compgen -f -- "${cur}") )
    fi
}

complete -F _tengo_completion tengo
`

const zshCompletion = `#compdef tengo
# zsh completion for tengo
# Add to a directory in $fpath, e.g.:
#   tengo -completion zsh > ~/.zfunc/_tengo
#   autoload -Uz compinit && compinit

_tengo() {
    _arguments \
        '-F[Find occurrences of text]:text:' \
        '-R[Replacement text (requires -F)]:text:' \
        '-validate[Validate file syntax and exit]' \
        '-format[Auto-format file in place]' \
        '-L[Print line n and exit]:line number:' \
        '-key[Read value at dot-notation key path]:path:' \
        '-set[Set value at dot-notation key path]:path=value:' \
        '-type[File type for stdin input]:type:(yaml yml json toml xml ini)' \
        '-version[Print version and exit]' \
        '-update[Download and install the latest release]' \
        '-dry-run[Show diff without writing to disk]' \
        '-stdout[Print result to stdout instead of saving]' \
        '-backup[Write .bak backup before modifying]' \
        '-q[Suppress output; use exit codes only]' \
        '-regex[Treat -F pattern as a regular expression]' \
        '-confirm-all[Apply all replacements without prompts]' \
        '-completion[Print shell completion script]:shell:(bash zsh fish)' \
        '*:file:_files'
}

_tengo
`

const fishCompletion = `# fish completion for tengo
# Install:
#   tengo -completion fish > ~/.config/fish/completions/tengo.fish

complete -c tengo -s F        -d 'Find occurrences of text'                    -r
complete -c tengo -s R        -d 'Replacement text (requires -F)'               -r
complete -c tengo -l validate -d 'Validate file syntax and exit'
complete -c tengo -l format   -d 'Auto-format file in place'
complete -c tengo -s L        -d 'Print line n and exit'                        -r
complete -c tengo -l key      -d 'Read value at dot-notation key path'          -r
complete -c tengo -l set      -d 'Set value at dot-notation key path'           -r
complete -c tengo -l type     -d 'File type for stdin input'                    -r -a 'yaml yml json toml xml ini'
complete -c tengo -l version  -d 'Print version and exit'
complete -c tengo -l update   -d 'Download and install the latest release'
complete -c tengo -l dry-run  -d 'Show diff without writing to disk'
complete -c tengo -l stdout   -d 'Print result to stdout instead of saving'
complete -c tengo -l backup   -d 'Write .bak backup before modifying'
complete -c tengo -s q        -d 'Suppress output; use exit codes only'
complete -c tengo -l regex    -d 'Treat -F pattern as a regular expression'
complete -c tengo -l confirm-all -d 'Apply all replacements without prompts'
complete -c tengo -l completion  -d 'Print shell completion script'             -r -a 'bash zsh fish'
`
