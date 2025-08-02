#!/bin/bash

# Script para ejecutar tests de memoria
# Uso: ./run_test.sh <numero_test>

show_help() {
    echo "Uso: $0 <numero_test>"
    echo "Tests disponibles: 1-6"
}

if [ $# -eq 0 ] || ! [[ "$1" =~ ^[1-6]$ ]]; then
    show_help
    exit 1
fi

TEST_NUM=$1
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

case $TEST_NUM in
    1) CONFIG_FILE="$SCRIPT_DIR/config/1_test_PCP.json" ;;
    2) CONFIG_FILE="$SCRIPT_DIR/config/2_test_PLYMP.json" ;;
    3) CONFIG_FILE="$SCRIPT_DIR/config/3_test_SWAP.json" ;;
    4) CONFIG_FILE="$SCRIPT_DIR/config/4_test_CACHE.json" ;;
    5) CONFIG_FILE="$SCRIPT_DIR/config/5_test_TLB.json" ;;
    6) CONFIG_FILE="$SCRIPT_DIR/config/6_test_GEN.json" ;;
esac

echo "Ejecutando Memoria - Test $TEST_NUM"
go run memoria.go "$CONFIG_FILE"