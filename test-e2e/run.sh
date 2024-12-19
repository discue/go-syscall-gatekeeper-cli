#!/bin/bash

set -uo pipefail

main_path=""
current_dir=$(pwd)
exitCode=0

while true; do
    if [[ -f "$current_dir/main.go" ]]; then
        main_path="$current_dir"
        break
    fi
    current_dir=$(dirname $current_dir)
done

trap 'rm -rf .tmp || true' EXIT # Fixed trap syntax

# Find all files ending with .sh in all subfolders and then iterate over them
for file in $(find . -mindepth 2 -type f -name "*.sh"); do
    mkdir .tmp
    
    output=$("timeout" "20s" "$file" "$main_path" 2>&1)
    exitCode=$?
    
    if [ $exitCode -eq 0 ]; then
        echo -e "\e[32mok\033[0m $file "
        elif [ $exitCode -eq 124 ]; then
        echo -e "\e[33mtimeout\033[0m $file"
        echo -e "$output"
        exitCode=1
    else
        echo -e "\e[91mfailed\033[0m $file"
        echo -e ">> Expected exit code 0 and got $exitCode"
        echo -e "$output"
        exitCode=1
    fi
    
    if [[ -d ".tmp" ]]; then
        rm -rf ".tmp"
    fi
    
done

exit $exitCode