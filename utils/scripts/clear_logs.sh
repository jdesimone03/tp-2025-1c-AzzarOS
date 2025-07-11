#!/bin/bash

# Script to delete all .log files in the current directory and subdirectories

echo "Deleting all .log files..."

find . -type f -name "*.log" -delete

echo "All .log files have been deleted."