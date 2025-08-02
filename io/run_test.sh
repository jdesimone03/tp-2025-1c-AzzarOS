#!/bin/bash

# Script para ejecutar tests de IO
# Uso: ./run_test.sh <numero_test> [io_id]

show_help() {
    echo "Uso: $0 <numero_test> [io_id]"
    echo "Tests disponibles: 1-6"
    echo "IO IDs: 1, 2, 3, 4 (para test 6)"
}

if [ $# -eq 0 ] || ! [[ "$1" =~ ^[1-6]$ ]]; then
    show_help
    exit 1
fi

TEST_NUM=$1
IO_ID=${2:-"1"}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

ARGS=("DISCO")

# Solo algunos tests tienen archivos de config específicos
case $TEST_NUM in
    1) 
        if [ "$IO_ID" = "2" ]; then
            ARGS+=("$SCRIPT_DIR/config/1_test_PCP_2.json")
        fi
        ;;
    6) ARGS+=("$SCRIPT_DIR/config/6_test_GEN_${IO_ID}.json") ;;
esac

echo "Ejecutando IO${IO_ID} - Test $TEST_NUM"
go run io.go "${ARGS[@]}"