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

// Peticiones

type HandshakeIO struct {
	Nombre		string
	Interfaz	Interfaz
}

type PeticionCPU struct {
	Identificador	string
	CPU				CPU
}

type PeticionIO struct {
	PID            	uint
	NombreIfaz		string
	SuspensionTime 	int
}

// Utilidades

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






//------------------------- IDEAS AGARRADAS DE LOS PELOS (NI IDEA QUE HICIMOS) -------------------------
type Instruccion struct {
	Instruccion	string
	Argumentos	[]string
}

func DecodificarInstruccion(instruccion string) (string, error) {
	//leo primera instruccion
	//listaDeInstrucciones []Instruccion
	//leo noop -> ListaDeInstrucciones[0]
	return "", nil
}


/*
type PCB struct {
	PID            uint
	PC             uint
	Estado         string
	Instrucciones  []Instruccion
	MetricasConteo map[string]int
	MetricasTiempo map[string]int64
}
*/

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

type IoInstruction struct {
    Duration int 
	Nombre   string
}

type InitProcInstruction struct {
    ProcessName string
    MemorySize  int 
}

type DumpMemoryInstruction struct{}

type ExitInstruction struct{}