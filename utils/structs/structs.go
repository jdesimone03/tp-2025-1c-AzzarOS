package structs

type Interfaz struct {
	IP		string
	Puerto	int
}

type PeticionIO struct {
	Nombre		string
	Interfaz	Interfaz
}

type PeticionKernel struct {
	PID            	uint
	NombreIfaz		string
	SuspensionTime 	int
}

type PCB struct {
	PID            uint
	PC             uint
	Estado         string
	MetricasConteo map[string]int
	MetricasTiempo map[string]int64
}

const (
	EstadoNew     = "NEW"
	EstadoReady   = "READY"
	EstadoExec    = "EXEC"
	EstadoBlocked = "BLOCKED"
	EstadoExit    = "EXIT"
	EstadoWaiting = "SUSP_BLOCKED"
	EstadoRunning = "SUSP_READY"
)