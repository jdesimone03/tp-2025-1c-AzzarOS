package utils

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
	IPMemory           string `json:"ip_memory"`
	PortMemory         int    `json:"port_memory"`
	IPKernel           string `json:"ip_kernel"`
	PortKernel         int    `json:"port_kernel"`
	SchedulerAlgorithm string `json:"scheduler_algorithm"`
	NewAlgorithm       string `json:"new_algorithm"`
	Alpha              string `json:"alpha"`
	SuspensionTime     int    `json:"suspension_time"`
	LogLevel           string `json:"log_level"`
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
}

type PCB struct {
	PID             uint
	PC              uint
	Estado          string
	MetricasConteo  map[string]int
	MetricasTiempo  map[string]int64
}

const (
	EstadoNew     = "NEW"
	EstadoReady   = "READY"
	EstadoExec    = "EXEC"
	EstadoBlocked = "BLOCKED"
	EstadoExit    = "EXIT"
	EstadoWaiting = "WAITING"
	EstadoRunning = "RUNNING"
)

func NuevoPCB(pid uint) *PCB {
	return &PCB{
		PID:            pid,
		PC:             0,
		Estado:         EstadoNew,
		MetricasConteo: inicializarConteo(),
		MetricasTiempo: inicializarTiempo(),
	}
}

func inicializarConteo() map[string]int {
	return map[string]int{
		EstadoNew:     0,
		EstadoReady:   0,
		EstadoExec:    0,
		EstadoBlocked: 0,
		EstadoExit:    0,
		EstadoWaiting: 0,
		EstadoRunning: 0,
	}
}

func inicializarTiempo() map[string]int64 {
	return map[string]int64{
		EstadoNew:     0,
		EstadoReady:   0,
		EstadoExec:    0,
		EstadoBlocked: 0,
		EstadoExit:    0,
		EstadoWaiting: 0,
		EstadoRunning: 0,
	}
}