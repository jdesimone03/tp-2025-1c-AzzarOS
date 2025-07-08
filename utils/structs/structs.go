package structs

import (
	"slices"
	"sync"
)

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

type Metricas struct {
	AccesosATablas           uint
	InstruccionesSolicitadas uint
	BajadasASWAP             uint
	SubidasAMemoria          uint
	Lecturas                 uint
	Escrituras               uint
}

// --------------------------------- Estructuras de Instancias --------------------------------- //

type InstanciaCPU struct {
	IP         string
	Puerto     string
	Ejecutando bool
	PID        uint
}

type InterfazIO struct {
	IP     string
	Puerto string
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

type NoopInstruction struct{}

type WriteInstruction struct {
	LogicAddress int
	Data    string
	PID     uint
}

type ReadInstruction struct {
	Address int
	Size    int
	PID    uint
}

type GotoInstruction struct {
	TargetAddress int
}

// Syscalls

type IoInstruction struct {
	NombreIfaz     string
	SuspensionTime int
}

type InitProcInstruction struct {
	ProcessPath string
	MemorySize  int
}

type DumpMemoryInstruction struct{}

type ExitInstruction struct{}

// --------------------------------- Estructuras seguras --------------------------------- //
// --------------------------------- MAPS --------------------------------- //
type MapSeguro[K comparable, V any] struct {
	Map   map[K]V
	Mutex sync.Mutex
}

func NewMapSeguro[K comparable, V any]() *MapSeguro[K, V] {
	return &MapSeguro[K, V]{Map: make(map[K]V)}
}

func (ms *MapSeguro[K, V]) Agregar(key K, value V) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	ms.Map[key] = value
}

func (ms *MapSeguro[K, V]) Obtener(key K) (V, bool) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	value, ok := ms.Map[key]
	return value, ok
}

func (ms *MapSeguro[K, V]) Eliminar(key K) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	delete(ms.Map, key)
}

// --------------------------------- MAP CPU --------------------------------- //

type MapCPU MapSeguro[string, InstanciaCPU]

func NewMapCPU() *MapCPU {
	return &MapCPU{Map: make(map[string]InstanciaCPU)}
}

func (ms *MapCPU) Agregar(key string, value InstanciaCPU) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	ms.Map[key] = value
}

func (ms *MapCPU) Obtener(key string) (InstanciaCPU, bool) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	value, ok := ms.Map[key]
	return value, ok
}

func (ms *MapCPU) Ocupar(nombre string, pid uint) InstanciaCPU {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()

	cpu := ms.Map[nombre]
	cpu.Ejecutando = true
	cpu.PID = pid
	ms.Map[nombre] = cpu

	return cpu
}

func (ms *MapCPU) Liberar(pid uint) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()

	nombreCPU := BuscarCPUPorPID(ms.Map, pid)
	cpu := ms.Map[nombreCPU]
	cpu.Ejecutando = false
	ms.Map[nombreCPU] = cpu
}

func (ms *MapCPU) BuscarCPUDisponible() (string, bool) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()

	for nombre, cpu := range ms.Map {
		if !cpu.Ejecutando {
			return nombre, true
		}
	}
	return "", false
}

func (ms *MapCPU) BuscarCPUPorPID(pid uint) string {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	return BuscarCPUPorPID(ms.Map, pid)
}

func (ms *MapCPU) ObtenerPID(nombre string) uint {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	return ms.Map[nombre].PID
}

// Funcion sin mutex para uso interno
func BuscarCPUPorPID(ms map[string]InstanciaCPU, pid uint) string {
	for nombre, cpu := range ms {
		if cpu.PID == pid {
			return nombre
		}
	}
	return ""
}


// --------------------------------- SLICE MAP --------------------------------- //
type SliceMapSeguro MapSeguro[string, []EjecucionIO]

func NewSliceMapSeguro() *SliceMapSeguro {
	return &SliceMapSeguro{Map: make(map[string][]EjecucionIO)}
}

func (sms *SliceMapSeguro) Agregar(key string, value EjecucionIO) {
	sms.Mutex.Lock()
	defer sms.Mutex.Unlock()
	sms.Map[key] = append(sms.Map[key], value)
}

func (sms *SliceMapSeguro) Obtener(key string) ([]EjecucionIO, bool) {
	sms.Mutex.Lock()
	defer sms.Mutex.Unlock()
	value, ok := sms.Map[key]
	return value, ok
}

func (sms *SliceMapSeguro) ObtenerPrimero(key string) EjecucionIO {
	sms.Mutex.Lock()
	defer sms.Mutex.Unlock()
	slice, _ := sms.Map[key]
	return slice[0]
}

func (sms *SliceMapSeguro) EliminarPrimero(key string) EjecucionIO {
	sms.Mutex.Lock()
	defer sms.Mutex.Unlock()
	if slice, ok := sms.Map[key]; ok && len(slice) > 0 {
		primerElemento := slice[0]
		sms.Map[key] = slice[1:]
		return primerElemento
	}
	return EjecucionIO{} // Devolver un valor vacío o un error si la clave no existe o el slice está vacío
}

func (sms *SliceMapSeguro) BorrarLista(key string) {
	sms.Mutex.Lock()
	defer sms.Mutex.Unlock()
	delete(sms.Map, key)
}

func (sms *SliceMapSeguro) Longitud(key string) int {
	sms.Mutex.Lock()
	defer sms.Mutex.Unlock()
	return len(sms.Map[key])
}

func (sms *SliceMapSeguro) NoVacia(key string) bool {
	sms.Mutex.Lock()
	defer sms.Mutex.Unlock()
	return len(sms.Map[key]) > 0
}

// --------------------------------- COLAS --------------------------------- //
type ColaSegura struct {
	Cola  []PCB
	Mutex sync.Mutex
}

func NewColaSegura() *ColaSegura {
	return &ColaSegura{Cola: make([]PCB, 0)}
}

func (cs *ColaSegura) Agregar(pcb PCB) {
	cs.Mutex.Lock()
	cs.Cola = append(cs.Cola, pcb)
	cs.Mutex.Unlock()
}

func (cs *ColaSegura) Eliminar(indice int) {
	cs.Mutex.Lock()
	cs.Cola = slices.Delete(cs.Cola, indice, indice+1)
	cs.Mutex.Unlock()
}

func (cs *ColaSegura) Obtener(indice int) PCB {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()
	return cs.Cola[indice]
}

func (cs *ColaSegura) Buscar(pid uint) (PCB, int) {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()
	for i, pcb := range cs.Cola {
		if pcb.PID == pid {
			return pcb, i
		}
	}
	return PCB{}, -1
}

func (cs *ColaSegura) Actualizar(pid uint, nuevoPC uint) {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()
	for i, pcb := range cs.Cola {
		if pcb.PID == pid {
			cs.Cola[i].PC = nuevoPC
			return
		}
	}
}

func (cs *ColaSegura) Longitud() int {
	return len(cs.Cola)
}

func (cs *ColaSegura) Vacia() bool {
	return !(len(cs.Cola) > 0)
}

// --------------------------------- Utilidades --------------------------------- //

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

// --------------------------------- Memoria --------------------------------- //

type TLB struct {
	Entradas    []EntradaTLB `json:"entradas"`
	MaxEntradas int          `json:"max_entradas"`
	Algoritmo   string       `json:"algoritmo"`
}

type EntradaTLB struct {
	NumeroPagina        int  `json:"numero_pagina"`
	NumeroFrame         int  `json:"numero_frame"`
	BitPresencia        bool `json:"bit_presencia"`    // Indica si el frame esta presente en memoria
	PID                 int  `json:"pid"`              // Identificador del proceso al que pertenece el frame
	Llegada int  `json:"instante_referencia"` // Marca el instante de referencia para LRU
	Referencia int  `json:"referencia"` // Marca el instante de referencia para LRU
}
type ProcesoEnSwap struct {
	PID uint `json:"pid"` // Identificador del proceso
	Paginas []string `json:"paginas"` // Lista de páginas del proceso
}
type Tabla struct {
	Punteros []*Tabla `json:"Tabla"`
	Valores  []int `json:"Valores"` // Valores de los frames ocupados por las paginas
}

type PaginaCache struct {
	NumeroFrame   int  `json:"numero_frame"`  // Numero de pagina
	NumeroPagina  int   `json:"numero_pagina"`  // Numero de pagina en la tabla de paginas
	BitPresencia  bool   `json:"bit_presencia"`	// Indica si el frame esta presente en memoria
	BitModificado bool   `json:"bit_modificado"`// Indica si el frame ha sido modificado
	BitDeUso      bool   `json:"bit_uso"`// Indica si el frame ha sido usado recientemente
	PID           int    `json:"pid"`// Identificador del proceso al que pertenece el frame
	Contenido     []byte `json:"contenido"`// Contenido de la pagina
}
type FrameInfo struct {
	EstaOcupado bool `json:"esta_ocupado"` // Indica si el frame está ocupado
	PID         uint `json:"pid"`          // Identificador del proceso al que pertenece el frame
}

type TablaDePaginas struct {
	PID      uint             `json:"pid"`     // Identificador del proceso
	Entradas []EntradaDeTabla `json:"paginas"` // Lista de páginas, Por numero de pagina: cada una con su bit de presencia y modificado y nro frame en memoria
}

type EntradaDeTabla struct {
	BitPresencia  bool `json:"bit_presencia"`
	BitModificado bool `json:"bit_modificado"`
	NumeroDeFrame int  `json:"numero_frame"`
	// PunteroATabla *TablaDePaginas `json:"puntero_a_tabla"`
}

type CuerpoSolicitud struct {
	PID uint `json:"PID"`
	PC  uint `json:"PC"`
}

type PedidoDeInicializacion struct {
	PID            uint   `json:"PID"`
	TamanioProceso uint   `json:"TAM"`
	Path           string `json:"PATH"`
}

type ConfigMemoria struct {
	CantNiveles      int `json:"cant_niveles"`
	EntradasPorTabla int `json:"entradas_por_tabla"`
	TamanioPagina    int `json:"tam_pagina"`
}

type CacheStruct struct {
	Paginas []PaginaCache 
	Algoritmo string 
	Clock int // dato para saber donde quedó la "aguja" del clock
}