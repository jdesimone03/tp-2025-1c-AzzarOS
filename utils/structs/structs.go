package structs

// --------------------------------- Estructuras Generales --------------------------------- //

type PCB struct {
	PID            uint
	PC             uint
	Estado         string
	MetricasConteo map[string]int
	MetricasTiempo map[string]int64
}

// Se manda a la memoria para inicializar el proceso
type NuevoProceso struct {
	PID           uint
	Instrucciones string //path al archivo de instrucciones
	Tamanio       int
}

// --------------------------------- Estructuras de Instancias --------------------------------- //

type InstanciaCPU struct {
	IP         string
	Puerto     int
	Ejecutando bool
	PID        uint
}

type InterfazIO struct {
	IP     string
	Puerto int
}

type EjecucionCPU struct {
	PID uint
	PC  uint
}

type EjecucionIO struct {
	PID      uint
	TiempoMs int
}

// --------------------------------- Estructuras de Handshakes --------------------------------- //

type HandshakeCPU struct {
	Identificador string
	CPU           InstanciaCPU
}

type HandshakeIO struct {
	Nombre   string
	Interfaz InterfazIO
}

// --------------------------------- Estructuras de Instrucciones --------------------------------- //
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

type Syscall interface {
	isSyscall()
}

type SyscallInstruction struct {
	PID         uint
	Instruccion Syscall
}

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

// Syscalls

type IOInstruction struct {
	NombreIfaz     string
	SuspensionTime int
}

func (IOInstruction) isSyscall() {}

type InitProcInstruction struct {
	ProcessPath string
	MemorySize  int
}

func (InitProcInstruction) isSyscall() {}

type DumpMemoryInstruction struct{}

func (DumpMemoryInstruction) isSyscall() {}

type ExitInstruction struct{}

func (ExitInstruction) isSyscall() {}

// --------------------------------- Utilidades --------------------------------- //

type Respuesta struct {
	Mensaje string
}

const (
	EstadoNew         = "NEW"
	EstadoReady       = "READY"
	EstadoExec        = "EXEC"
	EstadoBlocked     = "BLOCKED"
	EstadoExit        = "EXIT"
	EstadoSuspBlocked = "SUSP_BLOCKED"
	EstadoSuspReady   = "SUSP_READY"
)

type TiempoEjecucion struct {
	PID    uint
	Tiempo int64
}