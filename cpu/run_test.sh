#!/bin/bash

# Script para ejecutar tests de CPU
# Uso: ./run_test.sh <numero_test> [variante] [cpu_id]

show_help() {
    echo "Uso: $0 <numero_test> [variante] [cpu_id]"
    echo "Tests disponibles: 1-6"
    echo "CPU IDs: CPU1, CPU2, CPU3, CPU4 (default: CPU1)"
    echo ""
    echo "Variantes disponibles:"
    echo "  Test 4: clock-m, clock (default: clock)"
    echo "  Test 5: fifo, lru (default: fifo)"
    echo ""
    echo "Ejemplos:"
    echo "  $0 1               # Test 1 con CPU1"
    echo "  $0 4 clock-m       # Test 4 con algoritmo clock-m y CPU1"
    echo "  $0 5 lru           # Test 5 con algoritmo lru y CPU1"
    echo "  $0 1 fifo CPU2     # Test 1 con CPU2"
    echo "  $0 6 fifo CPU3     # Test 6 con CPU3"
}

if [ $# -eq 0 ] || ! [[ "$1" =~ ^[1-6]$ ]]; then
    show_help
    exit 1
fi

TEST_NUM=$1
VARIANT=$2
CPU_ID=$3

# Para tests 1 y 6, si se especifica CPU en el segundo parámetro, ajustar
if [[ $TEST_NUM == "1" || $TEST_NUM == "6" ]] && [[ $VARIANT =~ ^CPU[1-4]$ ]]; then
    CPU_ID=$VARIANT
    VARIANT=""
fi

# Defaults
CPU_ID=${CPU_ID:-"CPU1"}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Extraer número de CPU (CPU1 -> 1, CPU2 -> 2, etc.)
CPU_NUM=${CPU_ID#CPU}

case $TEST_NUM in
    1) CONFIG_FILE="$SCRIPT_DIR/config/${TEST_NUM}_test_PCP_${CPU_NUM}.json" ;;
    2|3) CONFIG_FILE="$SCRIPT_DIR/config/${TEST_NUM}_test_*.json" ;;
    4) 
        VARIANT=${VARIANT:-"clock"}
        case $VARIANT in
            "clock-m")
                CONFIG_FILE="$SCRIPT_DIR/config/4_test_CACHE_CLOCK_M.json"
                ;;
            "clock")
                CONFIG_FILE="$SCRIPT_DIR/config/4_test_CACHE_CLOCK.json"
                ;;
            *)
                echo "Error: Variante '$VARIANT' no válida para Test 4. Use: clock-m o clock"
                exit 1
                ;;
        esac
        ;;
    5)
        VARIANT=${VARIANT:-"fifo"}
        case $VARIANT in
            "fifo")
                CONFIG_FILE="$SCRIPT_DIR/config/5_test_TLB_FIFO.json"
                ;;
            "lru")
                CONFIG_FILE="$SCRIPT_DIR/config/5_test_TLB_LRU.json"
                ;;
            *)
                echo "Error: Variante '$VARIANT' no válida para Test 5. Use: fifo o lru"
                exit 1
                ;;
        esac
        ;;
    6) CONFIG_FILE="$SCRIPT_DIR/config/6_test_GEN_${CPU_NUM}.json" ;;
esac

# Resolver wildcard para tests 2-3
if [[ $CONFIG_FILE == *"*"* ]]; then
    CONFIG_FILE=$(ls $SCRIPT_DIR/config/${TEST_NUM}_test_*.json 2>/dev/null | head -1)
fi

# Verificar que el archivo de configuración existe
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: Archivo de configuración no encontrado: $CONFIG_FILE"
    exit 1
fi

echo "Ejecutando CPU $CPU_ID - Test $TEST_NUM${VARIANT:+ ($VARIANT)}"
go run cpu.go "$CPU_ID" "$CONFIG_FILE"