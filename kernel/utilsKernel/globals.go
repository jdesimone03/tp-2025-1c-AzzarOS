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
var TiempoEnColaBlocked = make(map[uint]*time.Timer)
var TiempoEnColaExecute = make(map[uint]int64)
var TiempoEstimado = make(map[uint]float64)

var contadorProcesos uint = 0

// scheduler_algorithm: LARGO plazo
// ready_ingress_algorithm: CORTO plazo

var NuevosProcesos = make(map[uint]structs.NuevoProceso)

var InstanciasCPU = make(map[string]structs.InstanciaCPU)
var Interfaces = make(map[string]structs.InterfazIO)

var ListaExecIO = structs.NewMapSeguro[structs.EjecucionIO]()
var ListaWaitIO = structs.NewMapSeguro[structs.EjecucionIO]()

var chCambioDeContexto = make(chan bool)
