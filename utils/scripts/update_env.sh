#!/usr/bin/env bash
# update_env.sh
# Uso: ./update_env.sh [-k IP] [-m IP] [-c IP] [-i IP] archivo.env

set -euo pipefail

# Variables para las nuevas IPs
new_memoria_ip=""
new_kernel_ip=""
new_cpu_ip=""
new_io_ip=""

# Parseo de argumentos
while [[ $# -gt 0 ]]; do
  case "$1" in
    --memoria|-m)
      shift
      new_memoria_ip="$1"
      ;;
    --kernel|-k)
      shift
      new_kernel_ip="$1"
      ;;
    --cpu|-c)
      shift
      new_cpu_ip="$1"
      ;;
    --io|-i)
      shift
      new_io_ip="$1"
      ;;
    --help|-h)
      echo "Uso: $0 [--memoria|-m IP] [--kernel|-k IP] [--cpu|-c IP] [--io|-i IP] archivo.env"
      exit 0
      ;;
    *)
      # Asumimos que el argumento es el archivo .env
      env_file="$1"
      ;;
  esac
  shift
done

# Validación del archivo
if [[ -z "${env_file-}" || ! -f "$env_file" ]]; then
  echo "Error: Debés indicar un archivo .env válido."
  exit 1
fi

# Reemplazo de IPs con sed
replace_ip() {
  local varname="$1"
  local newip="$2"
  if [[ -n "$newip" ]]; then
    sed -i -E "s|^(${varname}=).*|\1${newip}|" "$env_file"
    echo "✓ $varname actualizado a $newip"
  fi
}

# Aplicar cambios
replace_ip "IP_MEMORY" "$new_memoria_ip"
replace_ip "IP_KERNEL" "$new_kernel_ip"
replace_ip "IP_CPU"    "$new_cpu_ip"
replace_ip "IP_IO"     "$new_io_ip"

echo "Actualización completada en $env_file."
