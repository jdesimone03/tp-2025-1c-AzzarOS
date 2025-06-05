package structs

import "sync"

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

// --------------------------------- Estructuras seguras --------------------------------- //
type MapSeguro struct {
	Map map[string][]EjecucionIO
	Mutex sync.Mutex
}

func NewMapSeguro() *MapSeguro {
	return &MapSeguro{Map: make(map[string][]EjecucionIO)}
}

func (ms *MapSeguro) Agregar(key string, ejecucion EjecucionIO) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	ms.Map[key] = append(ms.Map[key], ejecucion)
}

func (ms *MapSeguro) Obtener(key string) ([]EjecucionIO, bool) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	slice, ok := ms.Map[key]
	return slice, ok
}

func (ms *MapSeguro) EliminarPrimero(key string) EjecucionIO {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	if slice, ok := ms.Map[key]; ok && len(slice) > 0 {
		primerElemento := slice[0]
		ms.Map[key] = slice[1:]
		return primerElemento
	}
	return EjecucionIO{} // Devolver un valor vacío o un error si la clave no existe o el slice está vacío
}

func (ms *MapSeguro) BorrarLista(key string) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	delete(ms.Map, key)

}

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