package logueador

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"reflect"
	"strings"
)

// CREA ARCHIVO .LOG
func ConfigurarLogger(nombreArchivoLog string) {
	logFile, err := os.OpenFile(nombreArchivoLog+".log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile)
	slog.Info("Logger " + nombreArchivoLog + ".log configurado")
}

// ---------------------------- Logs obligatorios ----------------------------//
// ---------------------------- KERNEL ----------------------------//
// Log obligatorio 1/8
func SyscallRecibida(pid uint, nombreSyscall string) {
	slog.Info(fmt.Sprintf("## (%d) - Solicitó syscall: %s", pid, nombreSyscall))
}

// Log obligatorio 2/8
func KernelCreacionDeProceso(pid uint) {
	slog.Info(fmt.Sprintf("## (%d) Se crea el proceso - Estado: NEW", pid))
}

// Log obligatorio 3/8
func CambioDeEstado(pid uint, estadoAnterior string, estadoNuevo string) {
	slog.Info(fmt.Sprintf("## (%d) pasa del estado %s al estado %s", pid, estadoAnterior, estadoNuevo))
}

// Log obligatorio 4/8
func MotivoDeBloqueo(pid uint, dispositivoIO string) {
	slog.Info(fmt.Sprintf("## (%d) - Bloqueado por IO: %s", pid, dispositivoIO))
}

// Log obligatorio 5/8
func KernelFinDeIO(pid uint) {
	slog.Info(fmt.Sprintf("## (%d) finalizó IO y pasa a READY", pid))
}

// Log obligatorio 6/8
func DesalojoSRT(pid uint) {
	slog.Info(fmt.Sprintf("## (%d) - Desalojado por algoritmo SJF/SRT", pid))
}

// Log obligatorio 7/8
func FinDeProceso(pid uint) {
	slog.Info(fmt.Sprintf("## (%d) - Finaliza el proceso", pid))
}

// Log obligatorio 8/8
func MetricasDeEstado(pid uint, metricasConteo map[string]int, metricasTiempo map[string]int64) {
	// Esto construye el string con todas las métricas
	var metricasString string
	for estado, conteo := range metricasConteo {
		tiempo := metricasTiempo[estado]
		metricasString += fmt.Sprintf("%s %d %d, ", estado, conteo, tiempo)
	}

	// Finalmente loguea las métricas
	slog.Info(fmt.Sprintf("## (%d) - Métricas de estado: %s", pid, metricasString))
}

// ---------------------------- MEMORIA ----------------------------//
// Log obligatorio 1/5
func MemoriaCreacionDeProceso(pid uint, tamanio int) {
	slog.Info(fmt.Sprintf("## PID: %d - Proceso Creado - Tamaño: %d", pid, tamanio))
}

// Log obligatorio 2/5
// Destrucción de Proceso: “## PID: <PID> - Proceso Destruido - Métricas - Acc.T.Pag: <ATP>; Inst.Sol.: <Inst.Sol.>; SWAP: <SWAP>; Mem.Prin.: <Mem.Prin.>; Lec.Mem.: <Lec.Mem.>; Esc.Mem.: <Esc.Mem.>”

// Log obligatorio 3/5
func ObtenerInstruccion(pid uint, pc uint, linea string) {
	slog.Info(fmt.Sprintf("## PID: %d - Obtener instrucción: %d - Instrucción: %s", pid, pc, linea))
}

// Log obligatorio 4/5
func OperacionEnEspacioDeUsuario(pid uint, accion string, direccionFisica int, tamanio int) {
	slog.Info(fmt.Sprintf("## PID: %d - %s - Dirección Física: %d - Tamaño: %d", pid, accion, direccionFisica, tamanio))
}

func EscrituraEnEspacioDeUsuario(pid uint, direccionFisica int, tamanio int) {
	OperacionEnEspacioDeUsuario(pid, "Escritura", direccionFisica, tamanio)
}

func LecturaEnEspacioDeUsuario(pid uint, direccionFisica int, tamanio int) {
	OperacionEnEspacioDeUsuario(pid, "Lectura", direccionFisica, tamanio)
}

// Log obligatorio 5/5
func MemoryDump(pid uint) {
	slog.Info(fmt.Sprintf("## PID: %d - Memory Dump solicitado", pid))
}

// ---------------------------- CPU ----------------------------//
// Log obligatorio 1/11
func FetchInstruccion(pid uint, pc uint) {
	slog.Info(fmt.Sprintf("## PID: %d - FETCH - Program Counter: %d", pid, pc))
}

// Log obligatorio 2/11
func InterrupcionRecibida() {
	slog.Info("## Llega interrupción al puerto Interrupt")
}

// Log obligatorio 3/11
func InstruccionEjecutada(pid uint, nombreInstruccion string, instruccionDecodificada any) {
	slog.Info(fmt.Sprintf("## PID: %d - Ejecutando: %s - %s", pid, nombreInstruccion, parametrosToString(instruccionDecodificada)))
}

// Log obligatorio 4/11
func LecturaEscrituraMemoria(pid uint, accion string, direccionFisica int, valor string) {
	slog.Info(fmt.Sprintf("## PID: %d - Acción: %s - Dirección Física: %d - Valor: %s", pid, accion, direccionFisica, valor))
}

func LecturaMemoria(pid uint, direccionFisica int, valor string) {
	LecturaEscrituraMemoria(pid, "Lectura", direccionFisica, valor)
}

func EscrituraMemoria(pid uint, direccionFisica int, valor string) {
	LecturaEscrituraMemoria(pid, "Escritura", direccionFisica, valor)
}

// Log obligatorio 5/11
func ObtenerMarco(pid uint, numeroPagina int, numeroMarco int) {
	slog.Info(fmt.Sprintf("## PID: %d - OBTENER MARCO - Página: %d - Marco: %d", pid, numeroPagina, numeroMarco))
}

// Log obligatorio 6/11
func TLBHit(pid uint, numeroPagina int) {
	slog.Info(fmt.Sprintf("## PID: %d - TLB HIT - Pagina: %d", pid, numeroPagina))
}

// Log obligatorio 7/11
func TLBMiss(pid uint, numeroPagina int) {
	slog.Info(fmt.Sprintf("## PID: %d - TLB MISS - Pagina: %d", pid, numeroPagina))
}

// Log obligatorio 8/11
func PaginaEncontradaEnCache(pid uint, numeroPagina int) {
	slog.Info(fmt.Sprintf("## PID: %d - Cache Hit - Pagina: %d", pid, numeroPagina))
}

// Log obligatorio 9/11
func PaginaFaltanteEnCache(pid uint, numeroPagina int) {
	slog.Info(fmt.Sprintf("## PID: %d - Cache Miss - Pagina: %d", pid, numeroPagina))
}

// Log obligatorio 10/11
func PaginaIngresadaEnCache(pid uint, numeroPagina int) {
	slog.Info(fmt.Sprintf("## PID: %d - Cache Add - Pagina: %d", pid, numeroPagina))
}

// Log obligatorio 11/11
func PaginaActualizadaDeCacheAMemoria(pid uint, numeroPagina int, frameEnMemoriaPrincipal int) {
	slog.Info(fmt.Sprintf("## PID: %d - Memory Update - Pagina: %d - Frame: %d", pid, numeroPagina, frameEnMemoriaPrincipal))
}

// ---------------------------- IO ----------------------------//
// Log obligatorio 1/2
func InicioIO(pid uint, tiempoMs int) {
	slog.Info(fmt.Sprintf("## PID: %d - Inicio de IO - Tiempo: %d", pid, tiempoMs))
}

// Log obligatorio 2/2
func FinalizacionIO(pid uint) {
	slog.Info(fmt.Sprintf("## PID: %d - Fin de IO", pid))
}

// ---------------------------- Utilidades ----------------------------//
func parametrosToString(instruccion any) string {
	v := reflect.ValueOf(instruccion)

	// Chequeamos que sea un struct
	if v.Kind() != reflect.Struct {
		return ""
	}

	var values []string
	for i := range v.NumField() {
		field := v.Type().Field(i)

		// Solo accedemos a los campos exportados
		if field.PkgPath == "" {
			val := v.Field(i).Interface()
			values = append(values, fmt.Sprintf("%v", val))
		}
	}

	return strings.Join(values, " ")
}
