#!/bin/bash
# Bash completion for claude-pipeline CLI

_pipeline_cli() {
    local cur prev words cword
    _init_completion || return

    local commands="skills tasks pipelines runs schedules templates status version help"
    local subcmds_skills="list get sync"
    local subcmds_tasks="list get create result cancel"
    local subcmds_pipelines="list get create run delete"
    local subcmds_runs="list get cancel"
    local subcmds_schedules="list get create trigger enable disable delete"

    # Find the command
    local cmd=""
    for ((i=1; i<${#words[@]}; i++)); do
        if [[ ${words[i]} != -* ]]; then
            cmd=${words[i]}
            break
        fi
    done

    case $cmd in
        skills)
            COMPREPLY=($(compgen -W "$subcmds_skills" -- "$cur"))
            ;;
        tasks)
            COMPREPLY=($(compgen -W "$subcmds_tasks" -- "$cur"))
            ;;
        pipelines)
            COMPREPLY=($(compgen -W "$subcmds_pipelines" -- "$cur"))
            ;;
        runs)
            COMPREPLY=($(compgen -W "$subcmds_runs" -- "$cur"))
            ;;
        schedules)
            COMPREPLY=($(compgen -W "$subcmds_schedules" -- "$cur"))
            ;;
        templates)
            COMPREPLY=($(compgen -W "list use" -- "$cur"))
            ;;
        "")
            COMPREPLY=($(compgen -W "$commands" -- "$cur"))
            ;;
    esac

    # File completion for certain commands
    if [[ $prev == "create" && $cmd == "pipelines" ]]; then
        _filedir json
    fi
}

complete -F _pipeline_cli pipeline-cli
complete -F _pipeline_cli claude-pipeline