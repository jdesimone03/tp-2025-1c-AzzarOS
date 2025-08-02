#!/bin/bash

# Script para ejecutar tests de kernel
# Uso: ./run_test.sh <numero_test> [variante]

show_help() {
    echo "Uso: $0 <numero_test> [variante]"
    echo "Tests disponibles:"
    echo "  1 [fifo|sjf|srt] - Test PCP (default: fifo)"
    echo "  2 [fifo|pmcp] - Test PLYMP (default: fifo)" 
    echo "  3 - Test SWAP"
    echo "  4 - Test CACHE"
    echo "  5 - Test TLB"
    echo "  6 - Test GEN"
    echo ""
    echo "Ejemplos:"
    echo "  $0 1 sjf"
    echo "  $0 2 pmcp"
}

if [ $# -eq 0 ] || ! [[ "$1" =~ ^[1-6]$ ]]; then
    show_help
    exit 1
fi

TEST_NUM=$1
VARIANT=${2:-"fifo"}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

case $TEST_NUM in
    1) 
        case $VARIANT in
            "fifo")
                ARGS=("PLANI_CORTO_PLAZO" "0" "$SCRIPT_DIR/config/1_test_PCP_FIFO.json")
                ;;
            "sjf")
                ARGS=("PLANI_CORTO_PLAZO" "0" "$SCRIPT_DIR/config/1_test_PCP_SJF.json")
                ;;
            "srt")
                ARGS=("PLANI_CORTO_PLAZO" "0" "$SCRIPT_DIR/config/1_test_PCP_SRT.json")
                ;;
            *)
                echo "Error: Variante '$VARIANT' no válida para Test 1. Use: fifo, sjf o srt"
                exit 1
                ;;
        esac
        ;;
    2) 
        case $VARIANT in
            "fifo")
                ARGS=("PLANI_LYM_PLAZO" "0" "$SCRIPT_DIR/config/2_test_PLYMP.json")
                ;;
            "pmcp")
                ARGS=("PLANI_LYM_PLAZO" "0" "$SCRIPT_DIR/config/2_test_PLYMP_PMCP.json")
                ;;
            *)
                echo "Error: Variante '$VARIANT' no válida para Test 2. Use: fifo o pmcp"
                exit 1
                ;;
        esac
        ;;
    3) ARGS=("MEMORIA_IO" "90" "$SCRIPT_DIR/config/3_test_SWAP.json") ;;
    4) ARGS=("MEMORIA_BASE" "256" "$SCRIPT_DIR/config/4_test_CACHE.json") ;;
    5) ARGS=("MEMORIA_BASE_TLB" "256" "$SCRIPT_DIR/config/5_test_TLB.json") ;;
    6) ARGS=("ESTABILIDAD_GENERAL" "0" "$SCRIPT_DIR/config/6_test_GEN.json") ;;
esac

echo "Ejecutando Kernel - Test $TEST_NUM (${VARIANT})"
go run kernel.go "${ARGS[@]}"