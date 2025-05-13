package structs

// Estructuras generales

type Interfaz struct {
	IP		string
	Puerto	int
}

type CPU struct {
	IP		string
	Puerto	int
}

type PCB struct {
	PID            uint
	PC             uint
	Estado         string
	MetricasConteo map[string]int
	MetricasTiempo map[string]int64
}

type Proceso struct {
	PID             uint
	Instrucciones   string //path al archivo de instrucciones
	Tamanio			int
}

// Peticiones

type HandshakeIO struct {
	Nombre		string
	Interfaz	Interfaz
}

type HandshakeCPU struct {
	Identificador	string
	CPU				CPU
}

type PeticionMemoria struct {
	PID            	uint
	PC			 	uint
}

// Utilidades

type Respuesta struct {
	Mensaje	string
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

type EsperaIO struct {
	PID			uint
	TiempoMs	int
}


//--------------------------------- PROPUESTA TOMI P ----------------------------------------------------
type InstructionType int

const (
    INST_UNKNOWN InstructionType = iota // Parta los valores desconocidos
    INST_NOOP
    INST_WRITE
    INST_READ
    INST_GOTO
    INST_IO
    INST_INIT_PROC
    INST_DUMP_MEMORY
    INST_EXIT
)

type NoopInstruction struct{}

type WriteInstruction struct {
    Address int
    Data    string
}

type ReadInstruction struct {
    Address int
    Size    int
}

type GotoInstruction struct {
    TargetAddress int
}

type IOInstruction struct {
	PID            	uint
	NombreIfaz		string
	SuspensionTime 	int
}

type InitProcInstruction struct {
    ProcessPath string
    MemorySize  int 
}

type DumpMemoryInstruction struct{}

type ExitInstruction struct{
	PID	uint
}