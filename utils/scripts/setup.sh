#!/bin/bash

echo "=== Configuración inicial del proyecto AzzarOS ==="
echo ""

# Hacer ejecutables todos los archivos .sh en el proyecto
echo "1. Configurando permisos de ejecución para scripts..."
find . -name "*.sh" -type f -exec chmod +x {} \;

echo "   ✓ Archivos .sh configurados:"
find . -name "*.sh" -type f | sed 's|^\./||' | sort

echo ""

# Configurar archivo .env
echo "2. Configurando archivo de entorno..."

# Buscar .env.example en el directorio raíz del proyecto
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_EXAMPLE="$PROJECT_ROOT/.env.example"
ENV_FILE="$PROJECT_ROOT/.env"

if [ -f "$ENV_EXAMPLE" ]; then
    if [ ! -f "$ENV_FILE" ]; then
        echo "   ✓ Copiando .env.example a .env..."
        cp "$ENV_EXAMPLE" "$ENV_FILE"
        echo "   ✓ Archivo .env creado"
    else
        echo "   ⚠ El archivo .env ya existe"
        read -p "   ¿Desea sobrescribirlo? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            cp "$ENV_EXAMPLE" "$ENV_FILE"
            echo "   ✓ Archivo .env sobrescrito"
        else
            echo "   - Manteniendo .env existente"
        fi
    fi
    
    echo ""
    echo "3. Abriendo nano para configurar variables de entorno..."
    echo "   (Presione Ctrl+X para guardar y salir)"
    
    # Usar nano como editor
    nano "$ENV_FILE"
    
else
    echo "   ⚠ No se encontró .env.example en $ENV_EXAMPLE"
    echo "   Creando .env vacío..."
    touch "$ENV_FILE"
    echo "   Configure manualmente las variables en: $ENV_FILE"
fi

echo ""
echo "=== Configuración completada ==="
echo ""
echo "Siguiente pasos:"
echo "- Revisar y ajustar las variables en .env si es necesario"
echo "- Ejecutar tests con: ./run_test.sh <modulo> <numero>"
echo ""