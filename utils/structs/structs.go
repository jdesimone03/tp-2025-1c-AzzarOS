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

type PeticionIO struct {
	Nombre		string
	Interfaz	Interfaz
}

type PeticionCPU struct {
	Identificador	string
	CPU				CPU
}

type PeticionKernel struct {
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








//------------------------- IDEAS AGARRADAS DE LOS PELOS (NI IDEA QUE HICIMOS) -------------------------
//
//type Instrucciones struct {
//	Noop 	   string
//	Write 	   map[string]Write
//	Read 	   string
//	Goto 	   string
//}
//
//type Write struct {
//	Direcccion string
//	Datos 	   []int
//}
//
//type Instruccion struct {
//	Instruccion	string
//	Argumentos	[]string
//}
//
//func DecodificarInstruccion(instruccion string) (string, error) {
//	//leo primera instruccion
//	//listaDeInstrucciones []Instruccion
//	//leo noop -> ListaDeInstrucciones[0]
//	return "", nil
//}
//
//
///*
//type PCB struct {
//	PID            uint
//	PC             uint
//	Estado         string
//	Instrucciones    []Instruccion
//	MetricasConteo map[string]int
//	MetricasTiempo map[string]int64
//}
//*/
//
//-------------------------------------------------------------------------------------