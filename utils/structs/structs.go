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
	Nombre	   string
	IP         string
	Puerto     string
}

type InterfazIO struct {
	Nombre     string
	IP         string
	Puerto     string
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
	Data         string
	PID          uint
}

type ReadInstruction struct {
	Address int
	Size    int
	PID     uint
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

// --------------------------------- MAP CPU EXEC --------------------------------- //

type MapCPUExec MapSeguro[string, EjecucionCPU]

func NewMapCPUExec() *MapCPUExec {
	return &MapCPUExec{Map: make(map[string]EjecucionCPU)}
}

func (ms *MapCPUExec) Agregar(key string, value EjecucionCPU) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	ms.Map[key] = value
}

func (ms *MapCPUExec) Obtener(key string) (EjecucionCPU, bool) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	value, ok := ms.Map[key]
	return value, ok
}

func (ms *MapCPUExec) Eliminar(key string) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()
	delete(ms.Map, key)
}

func (ms *MapCPUExec) Buscar(pid uint) (string, bool) {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()

	for k, v := range(ms.Map) {
		if v.PID == pid {
			return k, true
		}
	}
	return "", false
}

func (ms *MapCPUExec) BuscarYEliminar(pid uint) bool {
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()

	for k, v := range(ms.Map) {
		if v.PID == pid {
			delete(ms.Map, k)
			return true
		}
	}
	return false
}

// --------------------------------- MAP IO WAIT --------------------------------- //

type MapIOWait MapSeguro[string, []EjecucionIO]

func NewMapIOWait() *MapIOWait {
	return &MapIOWait{Map: make(map[string][]EjecucionIO)}
}

func (sms *MapIOWait) Agregar(key string, value EjecucionIO) {
	sms.Mutex.Lock()
	defer sms.Mutex.Unlock()
	sms.Map[key] = append(sms.Map[key], value)
}

func (sms *MapIOWait) EliminarPrimero(key string) EjecucionIO {
	sms.Mutex.Lock()
	defer sms.Mutex.Unlock()
	if slice, ok := sms.Map[key]; ok && len(slice) > 0 {
		primerElemento := slice[0]
		sms.Map[key] = slice[1:]
		return primerElemento
	}
	return EjecucionIO{}
}

func (sms *MapIOWait) NoVacia(key string) bool {
	sms.Mutex.Lock()
	defer sms.Mutex.Unlock()
	return len(sms.Map[key]) > 0
}

// --------------------------------- MAP IO EXEC --------------------------------- //

type MapIOExec MapSeguro[InterfazIO, []EjecucionIO]

func NewMapIOExec() *MapIOExec {
	return &MapIOExec{Map: make(map[InterfazIO][]EjecucionIO)}
}

func (sms *MapIOExec) Agregar(key InterfazIO, value EjecucionIO) {
	sms.Mutex.Lock()
	defer sms.Mutex.Unlock()
	sms.Map[key] = append(sms.Map[key], value)
}

func (sms *MapIOExec) Obtener(key InterfazIO) ([]EjecucionIO, bool) {
	sms.Mutex.Lock()
	defer sms.Mutex.Unlock()
	value, ok := sms.Map[key]
	return value, ok
}

func (sms *MapIOExec) EliminarPrimero(key InterfazIO) EjecucionIO {
	sms.Mutex.Lock()
	defer sms.Mutex.Unlock()
	if slice, ok := sms.Map[key]; ok && len(slice) > 0 {
		primerElemento := slice[0]
		sms.Map[key] = slice[1:]
		return primerElemento
	}
	return EjecucionIO{} // Devolver un valor vacío o un error si la clave no existe o el slice está vacío
}

// --------------------------------- MAP CHANNELS --------------------------------- //

// Estructura genérica para manejar maps de channels
type MapChannels[T any] struct {
    channels map[uint]chan T
    mutex    sync.RWMutex
}

func NewMapChannels[T any]() *MapChannels[T] {
    return &MapChannels[T]{
        channels: make(map[uint]chan T),
    }
}

func (mc *MapChannels[T]) ObtenerChannel(pid uint, bufferSize int) chan T {
    mc.mutex.Lock()
    defer mc.mutex.Unlock()
    
    if ch, existe := mc.channels[pid]; existe {
        return ch
    }
    
    // Crear nuevo channel si no existe
    ch := make(chan T, bufferSize)
    mc.channels[pid] = ch
    return ch
}

func (mc *MapChannels[T]) LimpiarChannel(pid uint) {
    mc.mutex.Lock()
    defer mc.mutex.Unlock()
    
    if ch, existe := mc.channels[pid]; existe {
        close(ch)
        delete(mc.channels, pid)
    }
}

func (mc *MapChannels[T]) Señalizar(pid uint, valor T) bool {
    mc.mutex.RLock()
    ch, existe := mc.channels[pid]
    mc.mutex.RUnlock()
    
    if !existe {
        return false
    }
    
    select {
    case ch <- valor:
        return true
    default:
        return false
    }
}

func (mc *MapChannels[T]) Existe(pid uint) bool {
    mc.mutex.RLock()
    defer mc.mutex.RUnlock()
    
    _, existe := mc.channels[pid]
    return existe
}

// --------------------------------- SLICES --------------------------------- //
type SliceSeguro[T any] struct {
	Cola  []T
	Mutex sync.Mutex
}

// --------------------------------- LISTA CPU --------------------------------- //

type ListaCPU SliceSeguro[InstanciaCPU]

func NewListaCPU() *ListaCPU {
	return &ListaCPU{Cola: make([]InstanciaCPU, 0)}
}

func (lc *ListaCPU) Agregar(elemento InstanciaCPU) {
	lc.Mutex.Lock()
	defer lc.Mutex.Unlock()
	lc.Cola = append(lc.Cola, elemento)
}

func (lc *ListaCPU) Obtener(indice int) InstanciaCPU {
	lc.Mutex.Lock()
	defer lc.Mutex.Unlock()
	value := lc.Cola[indice]
	return value
}

func (lc *ListaCPU) Eliminar(elemento InstanciaCPU) {
	lc.Mutex.Lock()
	defer lc.Mutex.Unlock()
	for i, interfaz := range lc.Cola {
		if interfaz == elemento {
			lc.Cola = slices.Delete(lc.Cola, i, i+1)
			return
		}
	}
}

func (lc *ListaCPU) Buscar(nombre string) (InstanciaCPU, bool) {
	lc.Mutex.Lock()
	defer lc.Mutex.Unlock()
	for _, interfaz := range lc.Cola {
		if interfaz.Nombre == nombre {
			return interfaz, true
		}
	}
	return InstanciaCPU{}, false
}

func (lc *ListaCPU) Longitud() int {
	lc.Mutex.Lock()
	defer lc.Mutex.Unlock()
	return len(lc.Cola)
}

// --------------------------------- LISTA INTERFACES --------------------------------- //

type ListaInterfaces SliceSeguro[InterfazIO]

func NewSliceSeguro[T any]() *ListaInterfaces {
	return &ListaInterfaces{Cola: make([]InterfazIO, 0)}
}

func (li *ListaInterfaces) Agregar(elemento InterfazIO) {
	li.Mutex.Lock()
	defer li.Mutex.Unlock()
	li.Cola = append(li.Cola, elemento)
}

func (li *ListaInterfaces) Obtener(indice int) InterfazIO {
	li.Mutex.Lock()
	defer li.Mutex.Unlock()
	value := li.Cola[indice]
	return value
}

func (li *ListaInterfaces) Eliminar(elemento InterfazIO) {
	li.Mutex.Lock()
	defer li.Mutex.Unlock()
	for i, interfaz := range li.Cola {
		if interfaz == elemento {
			li.Cola = slices.Delete(li.Cola, i, i+1)
			return
		}
	}
}

func (li *ListaInterfaces) Buscar(nombre string) (InterfazIO, bool) {
	li.Mutex.Lock()
	defer li.Mutex.Unlock()
	for _, interfaz := range li.Cola {
		if interfaz.Nombre == nombre {
			return interfaz, true
		}
	}
	return InterfazIO{}, false
}

func (li *ListaInterfaces) Longitud() int {
	li.Mutex.Lock()
	defer li.Mutex.Unlock()
	return len(li.Cola)
}

// --------------------------------- COLAS KERNEL --------------------------------- //
type ColaSegura SliceSeguro[PCB]

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

func (cs *ColaSegura) Actualizar(pid uint, nuevoPC uint) bool {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()
	for i, pcb := range cs.Cola {
		if pcb.PID == pid {
			cs.Cola[i].PC = nuevoPC
			return true
		}
	}
	return false
}

func (cs *ColaSegura) Longitud() int {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()
	return len(cs.Cola)
}

func (cs *ColaSegura) Vacia() bool {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()
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
	NumeroPagina int  `json:"numero_pagina"`
	NumeroFrame  int  `json:"numero_frame"`
	BitPresencia bool `json:"bit_presencia"`       // Indica si el frame esta presente en memoria
	PID          int  `json:"pid"`                 // Identificador del proceso al que pertenece el frame
	Llegada      int  `json:"instante_referencia"` // Marca el instante de referencia para LRU
	Referencia   int  `json:"referencia"`          // Marca el instante de referencia para LRU
}
type ProcesoEnSwap struct {
	PID     uint     `json:"pid"`     // Identificador del proceso
	Paginas []string `json:"paginas"` // Lista de páginas del proceso
}
type Tabla struct {
	Punteros []*Tabla `json:"Tabla"`
	Valores  []int    `json:"Valores"` // Valores de los frames ocupados por las paginas
}

type PaginaCache struct {
	NumeroFrame   int    `json:"numero_frame"`   // Numero de pagina
	NumeroPagina  int    `json:"numero_pagina"`  // Numero de pagina en la tabla de paginas
	BitPresencia  bool   `json:"bit_presencia"`  // Indica si el frame esta presente en memoria
	BitModificado bool   `json:"bit_modificado"` // Indica si el frame ha sido modificado
	BitDeUso      bool   `json:"bit_uso"`        // Indica si el frame ha sido usado recientemente
	PID           int    `json:"pid"`            // Identificador del proceso al que pertenece el frame
	Contenido     []byte `json:"contenido"`      // Contenido de la pagina
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
	Paginas   []PaginaCache
	Algoritmo string
	Clock     int // dato para saber donde quedó la "aguja" del clock
}
