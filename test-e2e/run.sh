#!/bin/bash

set -uo pipefail

fail_fast=0
only_pattern="*"

for i in "$@"; do # Quote "$@" to handle arguments with spaces correctly
    case "$i" in
        --fail-fast)
            fail_fast=1
        ;;
        --only=*)
            only_pattern="${i#*=}"
        ;;
        --only)
            # Check if there is an argument after --only
            if [[ -n "${@:$((${#@})):1}" ]]; then
                only_pattern="${@:$((${#@})):1}"
                shift # Consume the pattern argument
            else
                echo "When using --only, provide also a pattern to match test cases."
                exit 1
            fi
        ;;
        # Handle arguments that don't start with -- (the pattern itself) AFTER handling --only
        *)
            if [[ -z "$only_pattern" ]]; then # Check if we are expecting a pattern
                echo "Unknown option: $i"
                exit 1
            else # If --only was previously given and it does not have '=', consume the pattern
                only_pattern="$i"
            fi
        ;;
    esac
done

main_path=""
current_dir=$(pwd)
testExitCode=0
testFailures=()

while true; do
    if [[ -f "$current_dir/main.go" ]]; then
        main_path="$current_dir"
        break
    fi
    current_dir=$(dirname $current_dir)
done

trap 'rm -rf .tmp || true' EXIT # always clean up before exit

# Find all files ending with .sh in all subfolders and then iterate over them
for file in $(find . -mindepth 2 -type f -name "*.sh"); do
    if [[ "$only_pattern" != "*" && ! "$file" =~ $only_pattern ]]; then
        continue
        elif [[ "$file" =~ "before.sh" ]]; then
        continue
    fi
    
    mkdir .tmp
    
    dir=$(dirname $file)
    if [[ -f "$dir/before.sh" ]]; then
        "$dir/before.sh"
    fi
    
    echo -ne "\e[36mpending\033[0m $file\033[K\r"
    
    output=$("timeout" "20s" "$file" "$main_path" 2>&1)
    exitCode=$?
    
    echo -ne "\033[K\r" # reset previous log line to give the impression of overriding it
    
    if [ $exitCode -eq 0 ]; then
        echo -e "\e[32mok\033[0m $file "
        elif [ $exitCode -eq 124 ]; then
        echo -e "\e[33mtimeout\033[0m $file"
        echo -e "$output"
        testFailures+=("timeout -> $file")
        
        if [[ $fail_fast -eq 1 ]]; then
            break
        fi
    else
        echo -e "\e[91mfailed\033[0m $file"
        echo -e ">> Expected exit code 0 and got $exitCode"
        echo -e "$output"
        testFailures+=("failed -> $file")
        
        if [[ $fail_fast -eq 1 ]]; then
            break
        fi
    fi
    
    if [[ -d ".tmp" ]]; then
        rm -rf ".tmp"
    fi
done

echo "done"
# Print all filenames that failed
if [ ${#testFailures[@]} -gt 0 ]; then
    echo ""
    echo -e "You have \e[91m${#testFailures[@]}\033[0m test failure(s):"
    
    for file in "${testFailures[@]}"; do
        echo "$file"
    done
    
    echo ""
    echo "You can rerun single test cases running this script with \"--only ${testFailures[0]}\""
    echo ""
    echo "Exiting with 1"
    exit 1
fi

exit 0
