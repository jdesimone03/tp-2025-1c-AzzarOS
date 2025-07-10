package utilsKernel

import (
	"time"
	"utils/config"
	"utils/structs"
)

// ---------------------------- Variables globales ----------------------------//
// variables Config
var Config config.ConfigKernel

// Colas de estados de los procesos
var ColaNew = structs.NewColaSegura()
var ColaReady = structs.NewColaSegura()
var ColaExecute = structs.NewColaSegura()
var ColaBlocked = structs.NewColaSegura()
var ColaExit = structs.NewColaSegura()
var ColaSuspBlocked = structs.NewColaSegura()
var ColaSuspReady = structs.NewColaSegura()

// Map para trackear los timers de los procesos
var TiempoEnColaBlocked = structs.NewMapSeguro[uint, *time.Timer]()
var TiempoEnColaExecute = structs.NewMapSeguro[uint, int64]()
var TiempoEstimado = structs.NewMapSeguro[uint, float64]()

var contadorProcesos uint = 0

// scheduler_algorithm: LARGO plazo
// ready_ingress_algorithm: CORTO plazo

var NuevosProcesos = structs.NewMapSeguro[uint,structs.NuevoProceso]()
var ProcesosActuales = structs.NewMapSeguro[uint, structs.NuevoProceso]()
var ProcesosEnEspera = structs.NewMapSeguro[uint, structs.NuevoProceso]()

var InstanciasCPU = structs.NewMapCPU()
var Interfaces = structs.NewSliceSeguro[structs.InterfazIO]()

var ListaExecIO = structs.NewMapIOExec()
var ListaWaitIO = structs.NewMapIOWait()
