package logueador

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"reflect"
	"strconv"
	"strings"
	"utils/structs"
)

// ArchivoExiste verifica si un archivo ya existe en el directorio de trabajo
func ArchivoExiste(nombreArchivo string) bool {
	_, err := os.Stat(nombreArchivo + ".log")
	return !os.IsNotExist(err)
}

// CREA ARCHIVO .LOG
func ConfigurarLogger(nombreArchivoLog string, nivelLog string) {

	// Crear el directorio logs si no existe
    if _, err := os.Stat("logs"); os.IsNotExist(err) {
        err := os.MkdirAll("logs", 0755)
        if err != nil {
            panic(err)
        }
    }

	nombreCompleto := "logs/" + nombreArchivoLog
	i := 1
	for ArchivoExiste(nombreCompleto) {
		nombreCompleto = "logs/" + nombreArchivoLog + "_" + strconv.Itoa(i)
		i++
	}

	logFile, err := os.OpenFile(nombreCompleto+".log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	// Configurar el logger básico
	log.SetOutput(logFile)

	// Determinar el nivel de log
	var nivel slog.Level
	switch strings.ToUpper(nivelLog) {
	case "DEBUG":
		nivel = slog.LevelDebug
	case "INFO":
		nivel = slog.LevelInfo
	case "WARN", "WARNING":
		nivel = slog.LevelWarn
	case "ERROR":
		nivel = slog.LevelError
	default:
		nivel = slog.LevelInfo // Default
	}

	slog.SetLogLoggerLevel(nivel)

	log.SetFlags(log.Lmicroseconds)

	Info("Logger %s.log configurado", nombreArchivoLog)
}

// ---------------------------- Funciones log ----------------------------//
func Info(formato string, args ...any) {
	slog.Info(fmt.Sprintf(formato, args...))
}

func Error(formato string, args ...any) {
	slog.Error(fmt.Sprintf(formato, args...))
}

func Warn(formato string, args ...any) {
	slog.Warn(fmt.Sprintf(formato, args...))
}

func Debug(formato string, args ...any) {
	slog.Debug(fmt.Sprintf(formato, args...))
}

// ---------------------------- KERNEL ----------------------------//
// Log obligatorio 1/8
func SyscallRecibida(pid uint, nombreSyscall string) {
	Info("## (%d) - Solicitó syscall: %s", pid, nombreSyscall)
}

// Log obligatorio 2/8
func KernelCreacionDeProceso(pid uint) {
	Info("## (%d) Se crea el proceso - Estado: NEW", pid)
}

// Log obligatorio 3/8
func CambioDeEstado(pid uint, estadoAnterior string, estadoNuevo string) {
	Info("## (%d) pasa del estado %s al estado %s", pid, estadoAnterior, estadoNuevo)
}

// Log obligatorio 4/8
func MotivoDeBloqueo(pid uint, dispositivoIO string) {
	Info("## (%d) - Bloqueado por IO: %s", pid, dispositivoIO)
}

// Log obligatorio 5/8
func KernelFinDeIO(pid uint) {
	Info("## (%d) finalizó IO y pasa a READY", pid)
}

// Log obligatorio 6/8
func DesalojoSRT(pid uint) {
	Info("## (%d) - Desalojado por algoritmo SJF/SRT", pid)
}

// Log obligatorio 7/8
func FinDeProceso(pid uint) {
	Info("## (%d) - Finaliza el proceso", pid)
}

// Log obligatorio 8/8
func MetricasDeEstado(pcb structs.PCB) {
	// Esto construye el string con todas las métricas
	var metricasString string
	for estado, conteo := range pcb.MetricasConteo {
		tiempo := pcb.MetricasTiempo[estado]
		metricasString += fmt.Sprintf("[%s] %d accesos, tiempo %dms; ", estado, conteo, tiempo)
	}

	// Finalmente loguea las métricas
	Info("## (%d) - Métricas de estado: %s", pcb.PID, metricasString)
}

// ---------------------------- MEMORIA ----------------------------//
// Log obligatorio 1/5
func MemoriaCreacionDeProceso(pid uint, tamanio int) {
	Info("## PID: %d - Proceso Creado - Tamaño: %d", pid, tamanio)
}

// Log obligatorio 2/5
// Destrucción de Proceso: “## PID: <PID> - Proceso Destruido - Métricas - Acc.T.Pag: <ATP>; Inst.Sol.: <Inst.Sol.>; SWAP: <SWAP>; Mem.Prin.: <Mem.Prin.>; Lec.Mem.: <Lec.Mem.>; Esc.Mem.: <Esc.Mem.>”
func DestruccionDeProceso(pid int, metrica structs.Metricas) {
	Info("## PID: %d - Proceso Destruido - Métricas - Acc.T.Pag: %d; Inst.Sol.: %d; SWAP: %d; Mem.Prin.: %d; Lec.Mem.: %d; Esc.Mem.: %d", pid, metrica.AccesosATablas, metrica.InstruccionesSolicitadas, metrica.BajadasASWAP, metrica.SubidasAMemoria, metrica.Lecturas, metrica.Escrituras)
}

// Log obligatorio 3/5
func ObtenerInstruccion(pid uint, pc uint, linea string) {
	Info("## PID: %d - Obtener instrucción: %d - Instrucción: %s", pid, pc, linea)
}

// Log obligatorio 4/5
func OperacionEnEspacioDeUsuario(pid uint, accion string, direccionFisica int, tamanio int) {
	Info("## PID: %d - %s - Dirección Física: %d - Tamaño: %d", pid, accion, direccionFisica, tamanio)
}

func EscrituraEnEspacioDeUsuario(pid uint, direccionFisica int, tamanio int) {
	OperacionEnEspacioDeUsuario(pid, "Escritura", direccionFisica, tamanio)
}

func LecturaEnEspacioDeUsuario(pid uint, direccionFisica int, tamanio int) {
	OperacionEnEspacioDeUsuario(pid, "Lectura", direccionFisica, tamanio)
}

// Log obligatorio 5/5
func MemoryDump(pid uint) {
	Info("## PID: %d - Memory Dump solicitado", pid)
}

// ---------------------------- CPU ----------------------------//
// Log obligatorio 1/11
func FetchInstruccion(pid uint, pc uint) {
	Info("## PID: %d - FETCH - Program Counter: %d", pid, pc)
}

// Log obligatorio 2/11
func InterrupcionRecibida() {
	Info("## Llega interrupción al puerto Interrupt")
}

// Log obligatorio 3/11
func InstruccionEjecutada(pid uint, nombreInstruccion string, instruccionDecodificada any) {
	Info("## PID: %d - Ejecutando: %s - %s", pid, nombreInstruccion, parametrosToString(instruccionDecodificada))
}

// Log obligatorio 4/11
func LecturaEscrituraMemoria(pid uint, accion string, direccionFisica int, valor string) {
	Info("## PID: %d - Acción: %s - Dirección Física: %d - Valor: %s", pid, accion, direccionFisica, valor)
}

func LecturaMemoria(pid uint, direccionFisica int, valor string) {
	LecturaEscrituraMemoria(pid, "Lectura", direccionFisica, valor)
}

func EscrituraMemoria(pid uint, direccionFisica int, valor string) {
	LecturaEscrituraMemoria(pid, "Escritura", direccionFisica, valor)
}

// Log obligatorio 5/11
func ObtenerMarco(pid uint, numeroPagina int, numeroMarco int) {
	Info("## PID: %d - OBTENER MARCO - Página: %d - Marco: %d", pid, numeroPagina, numeroMarco)
}

// Log obligatorio 6/11
func TLBHit(pid uint, numeroPagina int) {
	Info("## PID: %d - TLB HIT - Pagina: %d", pid, numeroPagina)
}

// Log obligatorio 7/11
func TLBMiss(pid uint, numeroPagina int) {
	Info("## PID: %d - TLB MISS - Pagina: %d", pid, numeroPagina)
}

// Log obligatorio 8/11
func PaginaEncontradaEnCache(pid uint, numeroPagina int) {
	Info("## PID: %d - Cache Hit - Pagina: %d", pid, numeroPagina)
}

// Log obligatorio 9/11
func PaginaFaltanteEnCache(pid uint, numeroPagina int) {
	Info("## PID: %d - Cache Miss - Pagina: %d", pid, numeroPagina)
}

// Log obligatorio 10/11
func PaginaIngresadaEnCache(pid uint, numeroPagina int) {
	Info("## PID: %d - Cache Add - Pagina: %d", pid, numeroPagina)
}

// Log obligatorio 11/11
func PaginaActualizadaDeCacheAMemoria(pid uint, numeroPagina int, frameEnMemoriaPrincipal []byte) {
	Info("## PID: %d - Memory Update - Pagina: %d - Frame: %d", pid, numeroPagina, frameEnMemoriaPrincipal)
}

// ---------------------------- IO ----------------------------//
// Log obligatorio 1/2
func InicioIO(pid uint, tiempoMs int) {
	Info("## PID: %d - Inicio de IO - Tiempo: %d", pid, tiempoMs)
}

// Log obligatorio 2/2
func FinalizacionIO(pid uint) {
	Info("## PID: %d - Fin de IO", pid)
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
