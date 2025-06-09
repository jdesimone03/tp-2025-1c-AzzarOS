package utilsKernel

import (
	"bufio"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
	"utils"
	"utils/config"
	"utils/logueador"
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

// Semaforos mutex
var mxCambioPCB sync.Mutex
var mxExecIO sync.Mutex
var mxWaitIO sync.Mutex

var contadorProcesos uint = 0

// scheduler_algorithm: LARGO plazo
// ready_ingress_algorithm: CORTO plazo

var NuevosProcesos = make(map[uint]structs.NuevoProceso)

var InstanciasCPU = make(map[string]structs.InstanciaCPU)
var Interfaces = make(map[string]structs.InterfazIO)

var ListaExecIO = structs.NewMapSeguro[structs.EjecucionIO]()
var ListaWaitIO = structs.NewMapSeguro[structs.EjecucionIO]()

// ---------------------------- Handlers de endpoints ----------------------------//
func HandleHandshake(tipo string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch tipo {
		case "IO":
			interfaz, err := utils.DecodificarMensaje[structs.HandshakeIO](r)
			if err != nil {
				slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Inicializa la interfaz y el planificador
			Interfaces[interfaz.Nombre] = interfaz.Interfaz
			go PlanificadorIO(interfaz.Nombre)
			// MoverAExecIO(interfaz.Nombre)

			slog.Info(fmt.Sprintf("Nueva interfaz IO: %+v", interfaz))

		case "CPU":
			instancia, err := utils.DecodificarMensaje[structs.HandshakeCPU](r)
			if err != nil {
				slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			InstanciasCPU[instancia.Identificador] = instancia.CPU
			slog.Info(fmt.Sprintf("Nueva instancia CPU: %+v", instancia))

		default:
			slog.Error(fmt.Sprintf("FATAL: %s no es un modulo valido.", tipo))
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func HandleSyscall(tipo string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		rawPID := r.URL.Query().Get("pid")
		pid, _ := strconv.ParseUint(rawPID, 10, 32)

		// Log obligatorio 1/8
		logueador.SyscallRecibida(uint(pid), tipo)
		switch tipo {
		case "INIT_PROC":
			syscall, err := utils.DecodificarMensaje[structs.InitProcInstruction](r)
			if err != nil {
				slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			SyscallInitProc(uint(pid), *syscall)
		case "DUMP_MEMORY":
			syscall, err := utils.DecodificarMensaje[structs.DumpMemoryInstruction](r)
			if err != nil {
				slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			SyscallDumpMemory(uint(pid), *syscall)
		case "IO":
			syscall, err := utils.DecodificarMensaje[structs.IOInstruction](r)
			if err != nil {
				slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			SyscallIO(uint(pid), *syscall)
		case "EXIT":
			syscall, err := utils.DecodificarMensaje[structs.ExitInstruction](r)
			if err != nil {
				slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			SyscallExit(uint(pid), *syscall)
		default:
			slog.Error(fmt.Sprintf("FATAL: %s no es un tipo de syscall valida.", tipo))
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func GuardarContexto(w http.ResponseWriter, r *http.Request) {
	contexto, err := utils.DecodificarMensaje[structs.EjecucionCPU](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	slog.Info(fmt.Sprintf("(%d) Guardando contexto en PC: %d", contexto.PID, contexto.PC))

	// Desaloja las cpu que se estén usando.
	for nombre, instancia := range InstanciasCPU {
		if instancia.Ejecutando && instancia.PID == contexto.PID {
			instancia.Ejecutando = false
			InstanciasCPU[nombre] = instancia
		}
	}

	// Busca el proceso a guardar en la cola execute
	/* 	for i := range ColaExecute {
		if ColaExecute[i].PID == contexto.PID {
			ColaExecute[i].PC = contexto.PC
			break
		}
	} */

	w.WriteHeader(http.StatusOK)
}

func HandleIODisconnect(w http.ResponseWriter, r *http.Request) {
	ifaz, err := utils.DecodificarMensaje[string](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	slog.Warn(fmt.Sprintf("Se recibió notificación de desconexión de IO: %s", *ifaz))

	// Borra cualquier proceso que este ejecutando
	if ejecucion, existe := ListaExecIO.Obtener(*ifaz); existe {
		pid := ejecucion[0].PID
		Interrumpir(GetCPU(pid))
		MoverPCB(pid, ColaExecute, ColaExit, structs.EstadoExit)
		// Borro el proceso de la lista de ejecución
		ListaExecIO.EliminarPrimero(*ifaz)
	}

	if _, existe := Interfaces[*ifaz]; existe {
		delete(Interfaces, *ifaz)
	}

	w.WriteHeader(http.StatusOK)
}

func HandleIOEnd(w http.ResponseWriter, r *http.Request) {
	ifaz, err := utils.DecodificarMensaje[string](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ejecucion, existe := ListaExecIO.Obtener(*ifaz)
	if existe {
		pid := ejecucion[0].PID

		// Log obligatorio 5/8
		logueador.KernelFinDeIO(pid)

		// Borro el proceso de la lista de ejecución
		ListaExecIO.EliminarPrimero(*ifaz)
		MoverPCB(pid, ColaBlocked, ColaReady, structs.EstadoReady)
		w.WriteHeader(http.StatusOK)
	} else {
		slog.Error(fmt.Sprintf("No existe el proceso en la lista de ejecución: %s", *ifaz))
		w.WriteHeader(http.StatusBadRequest)
	}

}

// ---------------------------- Syscalls ----------------------------//
// No ejecuta directamente sino que lo encola en el planificador. El planificador despues tiene que ejecutarse al momento de iniciar la IO
func SyscallIO(pid uint, instruccion structs.IOInstruction) {

	Interrumpir(GetCPU(pid))

	nombre := instruccion.NombreIfaz
	tiempoMs := instruccion.SuspensionTime

	_, encontrada := Interfaces[nombre]
	if encontrada {
		espera := structs.EjecucionIO{
			PID:      pid,
			TiempoMs: tiempoMs,
		}
		lista, _ := ListaExecIO.Obtener(nombre)
		if len(lista) > 0 {
			// Enviar proceso a BLOCKED
			MoverPCB(pid, ColaExecute, ColaBlocked, structs.EstadoBlocked)

			// Iniciar timer de suspension
			IniciarTimerSuspension(pid)

			// Enviar proceso a ListaWaitIO
			ListaWaitIO.Agregar(nombre, espera)
		} else {
			// Enviar al proceso a ejecutar el IO
			MoverPCB(pid, ColaExecute, ColaBlocked, structs.EstadoBlocked)
			ListaExecIO.Agregar(nombre, espera)
		}
	} else {
		slog.Error(fmt.Sprintf("La interfaz %s no existe en el sistema", nombre))

		// Enviar proceso a EXIT
		MoverPCB(pid, ColaExecute, ColaExit, structs.EstadoExit)
	}
}

func SyscallDumpMemory(pid uint, instruccion structs.DumpMemoryInstruction) {
	// TODO
}

func SyscallInitProc(pid uint, instruccion structs.InitProcInstruction) {
	instrucciones := instruccion.ProcessPath
	tamanio := instruccion.MemorySize
	NuevoProceso(instrucciones, tamanio)
}

func SyscallExit(pid uint, instruccion structs.ExitInstruction) {
	// Seguir la logica de "Finalizacion de procesos"
	for nombre, instancia := range InstanciasCPU {
		if instancia.Ejecutando && instancia.PID == pid {
			instancia.Ejecutando = false
			InstanciasCPU[nombre] = instancia
		}
	}

	MoverPCB(pid, ColaExecute, ColaExit, structs.EstadoExit)
}

// ---------------------------- Planificadores ----------------------------//
func PlanificadorIO(nombre string) {
	for {
		interfaz, encontrada := Interfaces[nombre]
		if encontrada {
			if ListaExecIO.NoVacia(nombre) {
				// Enviar al IO el PID y el tiempo en ms
				proc := ListaExecIO.ObtenerPrimero(nombre)
				utils.EnviarMensaje(interfaz.IP, interfaz.Puerto, "ejecutarIO", proc)
			}
			if ListaWaitIO.NoVacia(nombre) {
				// Borra el primer elemento en la lista de espera
				aEjecutar := ListaWaitIO.EliminarPrimero(nombre)
				MoverPCB(aEjecutar.PID, ColaBlocked, ColaExecute, structs.EstadoExec)
				ListaExecIO.Agregar(nombre, aEjecutar)
			}
		} else {
			// Si se llega a desconectar el IO, se desconecta el planificador
			// Tengo que ver como hacer para que se borre la interfaz de la lista de interfaces al momento que se desconecta
			slog.Error(fmt.Sprintf("Interfaz %s no encontrada, desconectando el planificador", nombre))
			return
		}
	}
}

// Los procesos son creados con la syscall de INIT_PROC.
// Esta función solo los manda a ejecutar según el algoritmo de planificación.
func PlanificadorLargoPlazo() {
	slog.Info(fmt.Sprintf("Se cargara el siguiente algortimo para el planificador de largo plazo, %s", Config.SchedulerAlgorithm))
	var procesoAEnviar structs.NuevoProceso
	for {
		if ColaNew.NoVacia() {
			firstPCB := ColaNew.Obtener(0)
			switch Config.SchedulerAlgorithm {
			case "FIFO":
				procesoAEnviar = NuevosProcesos[firstPCB.PID]
				// Si no, no hace nada. Sigue con el bucle hasta que se libere
			case "PMCP":
				procesoMinimo := NuevosProcesos[firstPCB.PID]
				for _, pcb := range ColaNew.Cola {
					nuevoProceso := NuevosProcesos[pcb.PID]
					if nuevoProceso.Tamanio < procesoMinimo.Tamanio {
						procesoMinimo = nuevoProceso
					}
				}
				procesoAEnviar = procesoMinimo
			default:
				slog.Error(fmt.Sprintf("Algoritmo de planificacion de largo plazo no reconocido: %s", Config.SchedulerAlgorithm))
				return
			}
			slog.Info(fmt.Sprintf("Proceso a enviar - PID: %d, Archivo de Instrucciones: %s, Tamanio: %d", procesoAEnviar.PID, procesoAEnviar.Instrucciones, procesoAEnviar.Tamanio))
			// TODO Liberar map de nuevos procesos?
			respuesta := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "check-memoria", procesoAEnviar.Tamanio)
			if respuesta != "OK" {
				slog.Warn(fmt.Sprintf("(%d) No hay espacio en memoria para enviar el proceso. Esperando a que la memoria se libere...", procesoAEnviar.PID))
				// Implementar semaforos para que espere que termine un proceso
			}
			utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "nuevo-proceso", procesoAEnviar)
			MoverPCB(procesoAEnviar.PID, ColaNew, ColaReady, structs.EstadoReady)
			//TODO timesleep?
		}
	}
}

func PlanificadorCortoPlazo() {
	slog.Info(fmt.Sprintf("Se cargara el siguiente algortimo para el planificador de corto plazo, %s", Config.ReadyIngressAlgorithm))

	// var estimado = float64(Config.InitialEstimate)
	// var real int64
	for {
		for _, pcb := range ColaReady.Cola {
			nombreCPU, hayDisponible := GetCPUDisponible()
			if hayDisponible {
				switch Config.ReadyIngressAlgorithm {
				case "FIFO":
					ejecucion := structs.EjecucionCPU{
						PID: pcb.PID,
						PC:  pcb.PC,
					}

					// Marca como ejecutando
					cpu := InstanciasCPU[nombreCPU]
					cpu.Ejecutando = true
					cpu.PID = pcb.PID
					InstanciasCPU[nombreCPU] = cpu

					// Envia el proceso
					utils.EnviarMensaje(cpu.IP, cpu.Puerto, "dispatch", ejecucion)
					MoverPCB(pcb.PID, ColaReady, ColaExecute, structs.EstadoExec)
				case "SJF":
					// Est(n)=Estimado de la ráfaga anterior =
					// R(n) = Lo que realmente ejecutó de la ráfaga anterior en la CPU
					// Est(n+1) = El estimado de la próxima ráfaga
					// Est(n+1) =  alpha * R(n) + (1-alpha) * Est(n) ;    [0,1]
					// real = time.Now().UnixMilli() - TiempoEnColaExecute[pcb.PID]
					// estimadoSiguiente := EstimarRafaga(float64(estimado), float64(real))

					// if estimadoSiguiente > estimado {
					// 	estimado = estimadoSiguiente
					// 	//	desalojar ejecutando
					// 	//	mandar proceso a ejecutar
					// }

					//TODO IMPLEMENTAR
				case "SJF-SD":
					//1 estimar todos los procesos en la cola de ready
					//2 elegir el mas chico
					//3 mandar a ejecutar el mas chico
					//4 iniciar el timer
					//5 en base al ultimo timer reestimar todos los procesos en la cola de ready
				default:
					slog.Error(fmt.Sprintf("Algoritmo de planificacion de corto plazo no reconocido: %s", Config.ReadyIngressAlgorithm))
					return
				}
			}
		}
	}
}

func PlanificadorMedianoPlazo() {
	slog.Info("Iniciando Planificador de Mediano Plazo.")

	for {
		// Esto puede consumir más CPU. Considerar añadir un pequeño time.Sleep() si es necesario
		// para evitar el uso excesivo de CPU, por ejemplo: time.Sleep(10 * time.Millisecond)

		slog.Debug("PlanificadorMedianoPlazo: Ejecutando ciclo de verificación de suspensión.")

		// NOTA: Para un sistema robusto, el acceso concurrente a ColaBlocked y ProcesosEnTimer
		// desde múltiples goroutines (otros planificadores, handlers) debería protegerse con mutex.

		// Iterar sobre los PIDs que están actualmente en ColaBlocked.
		// Iteramos directamente sobre ColaBlocked, manejando la modificación del slice.

		for i := 0; i < ColaBlocked.Longitud(); {
			// Es crucial asegurar que el acceso a ColaBlocked[i] sea seguro si otras goroutines pueden modificarla.
			// La nota sobre el mutex es muy importante aquí.
			pcb := ColaBlocked.Obtener(i)
			currentPid := pcb.PID
			moved := false // Flag to track if the PCB was moved in this iteration

			if timer, timerExists := TiempoEnColaBlocked[currentPid]; timerExists {
				// Verificar si el timer ha expirado de forma no bloqueante.
				select {
				case <-timer.C: // El timer ha disparado.
					slog.Info(fmt.Sprintf("PlanificadorMedianoPlazo: Timer expirado para PID %d (en ColaBlocked).", currentPid))

					// Aquí se asume que el PCB con currentPid todavía está en ColaBlocked y es el que queremos mover.
					// MoverPCB buscará por PID.
					respuestaMemoria := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "mover-a-swap", currentPid)
					if respuestaMemoria != "OK" {
						slog.Error(fmt.Sprintf("PlanificadorMedianoPlazo: Error al mover el PCB con PID %d a swap: '%s'", currentPid, respuestaMemoria))
						// Si no se pudo mover a swap, no se mueve el PCB y se deja en ColaBlocked.
						break // No mover el PCB, continuar con el siguiente.
					}

					slog.Info(fmt.Sprintf("PlanificadorMedianoPlazo: Respuesta de 'mover-a-swap' para PID %d: '%s'", currentPid, respuestaMemoria))

					MoverPCB(currentPid, ColaBlocked, ColaSuspBlocked, structs.EstadoSuspBlocked)
					delete(TiempoEnColaBlocked, currentPid) // Eliminar el timer del mapa.
					moved = true                            // PCB fue movido, no se debe incrementar i.
				default:
					// Timer existe pero no ha expirado. No hacer nada con este PCB respecto al timer.
				}
			}

			if !moved {
				i++ // Incrementar el índice solo si el PCB actual no fue movido.
			}
			// Si moved == true, i no se incrementa, y el bucle procesará el nuevo elemento en el índice actual i.
		}
	}
}

func IniciarPlanificadores() {
	go PlanificadorCortoPlazo()
	go PlanificadorMedianoPlazo()
	go func() {
		slog.Info("Esperando confirmación para iniciar el planificador de largo plazo...")
		bufio.NewReader(os.Stdin).ReadBytes('\n') // espera al Enter
		go PlanificadorLargoPlazo()
	}()
}

// Mueve el pcb de una lista de procesos a otra EJ: mueve de NEW a READY y cambia al nuevo estado
func MoverPCB(pid uint, origen *structs.ColaSegura, destino *structs.ColaSegura, estadoNuevo string) {
	for i, pcb := range origen.Cola {
		if pcb.PID == pid {
			estadoActual := origen.Cola[i].Estado
			pcb.Estado = estadoNuevo // cambiar el estado del PCB

			// Log obligatorio 3/8
			logueador.CambioDeEstado(pid, estadoActual, estadoNuevo)

			pcb.MetricasConteo[estadoNuevo]++

			destino.Agregar(pcb)
			origen.Eliminar(i)

			return
		}
	}
}

// ---------------------------- Funciones de utilidad ----------------------------//
func NuevoProceso(rutaArchInstrucciones string, tamanio int) {
	proceso := structs.NuevoProceso{
		PID:           contadorProcesos, // PID actual
		Instrucciones: rutaArchInstrucciones,
		Tamanio:       tamanio,
	}

	//utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "nuevo-proceso", proceso)
	NuevosProcesos[proceso.PID] = proceso

	// Crea el PCB y lo inserta en NEW
	pcb := CrearPCB()
	ColaNew.Agregar(pcb)
	contadorProcesos++

	// Log obligatorio 2/8
	logueador.KernelCreacionDeProceso(pcb.PID)
}

func CrearPCB() structs.PCB {
	return structs.PCB{
		PID:            contadorProcesos,
		PC:             0,
		Estado:         structs.EstadoNew,
		MetricasConteo: make(map[string]int),
		MetricasTiempo: make(map[string]int64),
	}
}

func GetCPUDisponible() (string, bool) {
	for nombre, valores := range InstanciasCPU {
		if !valores.Ejecutando {
			return nombre, true
		}
	}
	return "", false
}

func GetCPU(pid uint) string {
	for nombre, valores := range InstanciasCPU {
		if valores.PID == pid {
			return nombre
		}
	}
	return ""
}

func EstimarRafaga(estimadoAnterior float64, realAnterior float64) float64 {
	return realAnterior*Config.Alpha + (1-Config.Alpha)*estimadoAnterior
}

func RecibirTiempoEjecucion(w http.ResponseWriter, r *http.Request) {
	tiempo, err := utils.DecodificarMensaje[structs.TiempoEjecucion](r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	TiempoEnColaExecute[tiempo.PID] = tiempo.Tiempo
	w.WriteHeader(http.StatusOK)
}

func Interrumpir(nombreCpu string) {
	cpu, existe := InstanciasCPU[nombreCpu]
	if !existe {
		slog.Error(fmt.Sprintf("No se pudo interrumpir %s ya que no existe en el sistema.", nombreCpu))
		return
	}

	url := fmt.Sprintf("http://%s:%d/interrupt", cpu.IP, cpu.Puerto)
	_, err := http.Get(url)

	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo interrumpir la CPU %s: %v", nombreCpu, err))
	}

}

func IniciarTimerSuspension(pid uint) {
	// Crear nuevo timer
	if _, exists := TiempoEnColaBlocked[pid]; exists {
		slog.Warn(fmt.Sprintf("Intento de iniciar un timer para PID %d que ya tiene uno activo. Ignorando.", pid))
		return
	}
	timer := time.NewTimer(time.Duration(Config.SuspensionTime) * time.Millisecond)
	TiempoEnColaBlocked[pid] = timer

	// Log de inicio del timer
	slog.Info(fmt.Sprintf("Timer de suspensión configurado para PID %d. Duración: %d ms. Será evaluado por PlanificadorMedianoPlazo.", pid, Config.SuspensionTime))
	// La lógica de expiración y movimiento ahora es manejada por PlanificadorMedianoPlazo.
}
