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

type NoopInstruction struct{}

type WriteInstruction struct {
	Address int
	Data    string
	PID     uint
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

type InitProcInstruction struct {
	ProcessPath string
	MemorySize  int
}

type DumpMemoryInstruction struct{}

type ExitInstruction struct{}

// --------------------------------- Estructuras seguras --------------------------------- //
// --------------------------------- MAPS --------------------------------- //
// Lo hice con tipo de dato generico por si alguna otra estructura lo necesitaba usar.
// Si resulta que es la unica lo sacamos para que sea la estructura que valga
type MapSeguro[T any] struct {
	Map   map[string][]T
	Mutex sync.Mutex
}

func NewMapSeguro[T any]() *MapSeguro[T] {
	return &MapSeguro[T]{Map: make(map[string][]T)}
}

func (ms *MapSeguro[T]) Agregar(key string, value T) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	ms.Map[key] = append(ms.Map[key], value)
}

func (ms *MapSeguro[T]) Obtener(key string) ([]T, bool) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	slice, ok := ms.Map[key]
	return slice, ok
}

func (ms *MapSeguro[T]) ObtenerPrimero(key string) T {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	slice, _ := ms.Obtener(key)
	return slice[0]
}

func (ms *MapSeguro[T]) EliminarPrimero(key string) T {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	if slice, ok := ms.Map[key]; ok && len(slice) > 0 {
		primerElemento := slice[0]
		ms.Map[key] = slice[1:]
		return primerElemento
	}
	var valorVacio T
	return valorVacio // Devolver un valor vacío o un error si la clave no existe o el slice está vacío
}

func (ms *MapSeguro[T]) BorrarLista(key string) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	delete(ms.Map, key)
}

func (ms *MapSeguro[T]) Longitud(key string) int {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	return len(ms.Map[key])
}

func (ms *MapSeguro[T]) NoVacia(key string) bool {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	return len(ms.Map[key]) > 0
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

func (cs *ColaSegura) NoVacia() bool {
	return len(cs.Cola) > 0
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

// --------------------------------- Memoria --------------------------------- //

type Tabla struct {
	Punteros []*Tabla
	Valores  []int
}

type PaginaCache struct {
	NumeroFrame   int    // Numero de pagina
	NumeroPagina  int    // Numero de pagina en la tabla de paginas
	BitPresencia  bool   // Indica si el frame esta presente en memoria
	BitModificado bool   // Indica si el frame ha sido modificado
	BitDeUso      bool   // Indica si el frame ha sido usado recientemente
	PID           int    // Identificador del proceso al que pertenece el frame
	Contenido     []byte // Contenido de la pagina
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
