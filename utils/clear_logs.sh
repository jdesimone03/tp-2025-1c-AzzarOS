#!/bin/bash

# Script to clear the content of all .log files in the current directory and subdirectories

echo "Clearing content of all .log files..."

find . -type f -name "*.log" -exec truncate -s 0 {} \;

echo "Content of all .log files has been cleared."