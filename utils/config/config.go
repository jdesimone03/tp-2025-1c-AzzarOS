package config

import (
	"bufio"
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"strings"
	"utils/logueador"
)

// ----------------------------------- CONFIGS --------------------------------------------------
type ConfigCPU struct {
	PortCPU          string `json:"port_cpu"`
	IPCPU            string `json:"ip_cpu"`
	IPMemory         string `json:"ip_memory"`
	PortMemory       string `json:"port_memory"`
	IPKernel         string `json:"ip_kernel"`
	PortKernel       string `json:"port_kernel"`
	TlbEntries       int    `json:"tlb_entries"`
	TlbReplacement   string `json:"tlb_replacement"`
	CacheEntries     int    `json:"cache_entries"`
	CacheReplacement string `json:"cache_replacement"`
	CacheDelay       int    `json:"cache_delay"`
	LogLevel         string `json:"log_level"`
}

type ConfigIO struct {
	IPKernel   string `json:"ip_kernel"`
	PortKernel string `json:"port_kernel"`
	PortIo     string `json:"port_io"`
	IPIo       string `json:"ip_io"`
	LogLevel   string `json:"log_level"`
}

type ConfigKernel struct {
	IPMemory              string  `json:"ip_memory"`
	PortMemory            string  `json:"port_memory"`
	IPKernel              string  `json:"ip_kernel"`
	PortKernel            string  `json:"port_kernel"`
	SchedulerAlgorithm    string  `json:"scheduler_algorithm"`
	ReadyIngressAlgorithm string  `json:"ready_ingress_algorithm"`
	Alpha                 float64 `json:"alpha"`
	InitialEstimate       int     `json:"initial_estimate"`
	SuspensionTime        int     `json:"suspension_time"`
	LogLevel              string  `json:"log_level"`
}

type ConfigMemory struct {
	PortMemory     string `json:"port_memory"`
	IPMemory       string `json:"ip_memory"`
	MemorySize     int    `json:"memory_size"`
	PageSize       int    `json:"page_size"`
	EntriesPerPage int    `json:"entries_per_page"`
	NumberOfLevels int    `json:"number_of_levels"`
	MemoryDelay    int    `json:"memory_delay"`
	SwapfilePath   string `json:"swapfile_path"`
	SwapDelay      int    `json:"swap_delay"`
	LogLevel       string `json:"log_level"`
	DumpPath       string `json:"dump_path"`
	ScriptsPath    string `json:"scripts_path"`
}

// ------------------------------------------------------------------------------------------------
func CargarConfiguracion(filePath string, configVar any) {
	CargarVariablesEntorno("../.env")

	file, err := os.Open(filePath)
	if err != nil {
		logueador.Error("No se pudo abrir el archivo de configuración  (%v)", err)
		panic(err)
	}
	defer file.Close()

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		logueador.Error("No se pudo leer el archivo de configuración (%v)", err)
		panic(err)
	}

	// Expand environment variables
	expandedContent := expandEnvWithMath(string(fileContent))

	// Decode the expanded JSON
	if err := json.Unmarshal([]byte(expandedContent), configVar); err != nil {
		logueador.Error("No se pudo decodificar el archivo JSON (%v)", err)
		panic(err)
	}

	logueador.Info("Configuración cargada correctamente: %+v", configVar)
}

// ----------------------------------------------- UTILIDADES .ENV --------------------------------------------
// Basicamente hice esto para facilitar las constantes de configuración comunes entre tests

func CargarVariablesEntorno(envPath string) {
	file, err := os.Open(envPath)
	if err != nil {
		logueador.Warn("No se pudo abrir el archivo .env (%v), usando variables de entorno del sistema", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first '=' only
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Only set if not already set in environment
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logueador.Error("Error leyendo archivo .env: %v", err)
	}
}

// Esto carga la variable de entorno si le metemos un operador de offset como + o -
// Ejemplo ${PORT_CPU+1000} = 8000 + 1000 = 9000
func expandEnvWithMath(content string) string {
	// Regex para capturar ${VAR+número} o ${VAR-número}
    re := regexp.MustCompile(`\$\{([^}]+)([+-]\d+)\}`)

    // Primero expandir variables con operaciones matemáticas
    content = re.ReplaceAllStringFunc(content, func(match string) string {
        // Extraer la variable y la operación
        parts := re.FindStringSubmatch(match)
        if len(parts) != 3 {
            return match
        }

        varName := parts[1]
        operation := parts[2]

        // Obtener el valor de la variable de entorno
        envValue := os.Getenv(varName)
        if envValue == "" {
            logueador.Warn("Variable de entorno %s no encontrada para operación %s", varName, operation)
            return match
        }

        // Convertir el valor base a número
        baseValue, err := strconv.Atoi(envValue)
        if err != nil {
            logueador.Warn("No se pudo convertir %s a número para la variable %s", envValue, varName)
            return match
        }

        // Extraer el operador y el valor
        operator := operation[0]
        operandStr := operation[1:]
        operand, err := strconv.Atoi(operandStr)
        if err != nil {
            logueador.Warn("No se pudo convertir %s a número en la operación %s", operandStr, operation)
            return match
        }

        // Realizar la operación
        var result int
        switch operator {
        case '+':
            result = baseValue + operand
        case '-':
            result = baseValue - operand
        default:
            logueador.Warn("Operador %c no soportado", operator)
            return match
        }

        return strconv.Itoa(result)
    })

    // Luego expandir variables normales (${VAR})
    return os.ExpandEnv(content)
}
