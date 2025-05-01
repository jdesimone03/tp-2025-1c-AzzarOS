package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

// ----------------------------------- CONFIGS --------------------------------------------------
type ConfigCPU struct {
	PortCPU          int    `json:"port_cpu"`
	IPCPU            string `json:"ip_cpu"`
	IPMemory         string `json:"ip_memory"`
	PortMemory       int    `json:"port_memory"`
	IPKernel         string `json:"ip_kernel"`
	PortKernel       int    `json:"port_kernel"`
	TlbEntries       int    `json:"tlb_entries"`
	TlbReplacement   string `json:"tlb_replacement"`
	CacheEntries     int    `json:"cache_entries"`
	CacheReplacement string `json:"cache_replacement"`
	CacheDelay       int    `json:"cache_delay"`
	LogLevel         string `json:"log_level"`
}

type ConfigIO struct {
	IPKernel   string `json:"ip_kernel"`
	PortKernel int    `json:"port_kernel"`
	PortIo     int    `json:"port_io"`
	IPIo       string `json:"ip_io"`
	LogLevel   string `json:"log_level"`
}

type ConfigKernel struct {
	IPMemory           		string `json:"ip_memory"`
	PortMemory         		int    `json:"port_memory"`
	IPKernel           		string `json:"ip_kernel"`
	PortKernel         		int    `json:"port_kernel"`
	SchedulerAlgorithm 		string `json:"scheduler_algorithm"`
	ReadyIngressAlgorithm	string `json:"ready_ingress_algorithm"`
	Alpha              		string `json:"alpha"`
	InitialEstimate			int	   `json:"initial_estimate"`
	SuspensionTime     		int    `json:"suspension_time"`
	LogLevel           		string `json:"log_level"`
}

type ConfigMemory struct {
	PortMemory     int    `json:"port_memory"`
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
	ScriptsPath	   string `json:"scripts_path"`
}

//------------------------------------------------------------------------------------------------
func CargarConfiguracion[T any](filePath string) *T {
	var config T

	file, err := os.Open(filePath)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo abrir el archivo de configuración  (%v)", err))
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el archivo JSON (%v)", err))
		panic(err)
	}

	slog.Info(fmt.Sprintf("Configuración cargada correctamente: %+v", config))
	return &config
}
