package utilsKernel

import (
	//"bufio"
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	//"os"
	"slices"
	"time"
	"utils"
	"utils/config"
	"utils/logueador"
	"utils/structs"
)

// ---------------------------- Variables globales ----------------------------//
// variables Config
var Config = config.CargarConfiguracion[config.ConfigKernel]("config.json")

// Colas de estados de los procesos
var ColaNew []structs.PCB
var ColaReady []structs.PCB
var ColaExecute []structs.PCB
var ColaBlocked []structs.PCB
var ColaExit []structs.PCB
var ColaSuspBlocked []structs.PCB
var ColaSuspReady []structs.PCB

// Map para trackear los timers de los procesos
var ProcesosEnTimer = make(map[uint]*time.Timer)

var contadorProcesos uint = 0

// scheduler_algorithm: LARGO plazo
// ready_ingress_algorithm: CORTO plazo

var InstanciasCPU = make(map[string]structs.CPU)
var Interfaces = make(map[string]structs.Interfaz)

var ListaExecIO = make(map[string][]structs.EsperaIO)
var ListaWaitIO = make(map[string][]structs.EsperaIO)

// ---------------------------- Handlers de endpoints ----------------------------//
func HandleHandshake(tipo string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch tipo {
		case "IO":
			interfaz, err := utils.DecodificarMensaje[structs.HandshakeIO](r)
			if err != nil {
				slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Error al decodificar mensaje"))
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
				w.Write([]byte("Error al decodificar mensaje"))
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
		// Log obligatorio 1/8
		logueador.SyscallRecibida(ColaExecute[0].PID, tipo)
		switch tipo {
		case "INIT_PROC":
			proceso, err := utils.DecodificarMensaje[structs.InitProcInstruction](r)
			if err != nil {
				slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			SyscallInitProc(*proceso)
		case "DUMP_MEMORY":
			return // No implementado
		case "IO":
			peticion, err := utils.DecodificarMensaje[structs.IOInstruction](r)
			if err != nil {
				slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			SyscallIO(*peticion)
		case "EXIT":
			proceso, err := utils.DecodificarMensaje[structs.ExitInstruction](r)
			if err != nil {
				slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			SyscallExit(*proceso)
		default:
			slog.Error(fmt.Sprintf("FATAL: %s no es un tipo de syscall valida.", tipo))
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)

	}
}

func GuardarContexto(w http.ResponseWriter, r *http.Request) {
	contexto, err := utils.DecodificarMensaje[structs.Ejecucion](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ColaExecute[0].PC = contexto.PC
	MoverPCB(contexto.PID, &ColaExecute, &ColaReady, structs.EstadoReady)

	w.WriteHeader(http.StatusOK)
}

// ---------------------------- Syscalls ----------------------------//
// No ejecuta directamente sino que lo encola en el planificador. El planificador despues tiene que ejecutarse al momento de iniciar la IO
func SyscallIO(peticion structs.IOInstruction) {
	pid := ColaExecute[0].PID
	nombre := peticion.NombreIfaz
	tiempoMs := peticion.SuspensionTime

	_, encontrada := Interfaces[nombre]
	if encontrada {
		if HandleIODisconnect(nombre) {
			return
		}
		espera := structs.EsperaIO{
			PID:      pid,
			TiempoMs: tiempoMs,
		}
		if len(ListaExecIO[nombre]) > 0 {
			// Enviar proceso a BLOCKED
			MoverPCB(pid, &ColaExecute, &ColaBlocked, structs.EstadoBlocked)

			// Iniciar timer de suspension
			IniciarTimerSuspension(pid)

			// Enviar proceso a ListaWaitIO
			ListaWaitIO[nombre] = append(ListaWaitIO[nombre], espera)
		} else {
			if HandleIODisconnect(nombre) {
				return
			}
			// Enviar al proceso a ejecutar el IO
			ListaExecIO[nombre] = append(ListaExecIO[nombre], espera)
		}
	} else {
		slog.Error(fmt.Sprintf("La interfaz %s no existe en el sistema", nombre))

		// Enviar proceso a EXIT
		MoverPCB(pid, &ColaExecute, &ColaExit, structs.EstadoExit)
	}
}

func SyscallInitProc(inst structs.InitProcInstruction) {
	instrucciones := inst.ProcessPath
	tamanio := inst.MemorySize
	NuevoProceso(instrucciones, tamanio)
}

func SyscallExit(proceso structs.ExitInstruction) {
	// Seguir la logica de "Finalizacion de procesos"
}

// ---------------------------- Planificadores ----------------------------//
func PlanificadorIO(nombre string) {
	for {
		interfaz, encontrada := Interfaces[nombre]
		if encontrada {
			lista := ListaExecIO[nombre]
			if len(lista) > 0 {
				// Enviar al IO el PID y el tiempo en ms
				proc := lista[0]
				if HandleIODisconnect(nombre) {
					LimpiarExecIO(nombre)
					return
				}
				// TODO investigar otra forma de hacer esto
				// Manejo del timeout
				timeoutMax := proc.TiempoMs + (proc.TiempoMs / 50) // Tiempo de espera maximo, es medio arbitrario que tiene que ser 50% mas del pedido. Se podria ajustar
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMax)*time.Millisecond)
				defer cancel()

				// Crea un canal que marca si termino la ejecución de IO
				done := make(chan bool, 1)
				go func() {
					utils.EnviarMensaje(interfaz.IP, interfaz.Puerto, "ejecutarIO", proc)
					done <- true
				}()

				select {
				case <-done:
					// SI TERMINA LA EJECUCION (FIN DE IO)
					// Borro el proceso de la lista de ejecucion
					LimpiarExecIO(nombre)

					// Log obligatorio 5/8
					logueador.KernelFinDeIO(proc.PID)
					MoverPCB(proc.PID, &ColaExecute, &ColaReady, structs.EstadoReady)

				case <-ctx.Done():
					// SI HAY DESCONEXION DE IO
					slog.Error(fmt.Sprintf("Timeout excedido para el proceso %d en la interfaz %s", proc.PID, nombre))
					MoverPCB(proc.PID, &ColaExecute, &ColaExit, structs.EstadoExit)
					delete(Interfaces, nombre)

					// Borro el proceso de la lista de ejecución
					LimpiarExecIO(nombre)
					return // Si se desconecta el io hay que desconectar el planificador
				}
			}
			aux := ListaWaitIO[nombre]
			if len(aux) > 0 {
				if HandleIODisconnect(nombre) {
					return
				}
				// Borra el primer elemento en la lista de espera
				aEjecutar := aux[0]
				aux := slices.Delete(aux, 0, 1)
				ListaWaitIO[nombre] = aux

				MoverPCB(aEjecutar.PID, &ColaBlocked, &ColaExecute, structs.EstadoExec)
				ListaExecIO[nombre] = append(ListaExecIO[nombre], aEjecutar)
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
	ejecutando := true
	slog.Info(fmt.Sprintf("Se cargara el siguiente algortimo para el planificador de largo plazo, %s", Config.SchedulerAlgorithm))
	for ejecutando {
		if len(ColaNew) > 0 {
			switch Config.SchedulerAlgorithm {
			case "FIFO":
				firstPCB := ColaNew[0]
				MoverPCB(firstPCB.PID, &ColaNew, &ColaReady, structs.EstadoReady)
				// Si no, no hace nada. Sigue con el bucle hasta que se libere
			case "PMCP":
				//ejecutar PMCP, no es de este checkpoint lo haremos despues (si dios quiere)
			default:
				slog.Error(fmt.Sprintf("Algoritmo de planificacion de largo plazo no reconocido: %s", Config.SchedulerAlgorithm))
				ejecutando = false
			}
		}
	}
	slog.Info("Finalizando planificador de largo plazo")
}

func PlanificadorCortoPlazo() {
	slog.Info(fmt.Sprintf("Se cargara el siguiente algortimo para el planificador de corto plazo, %s", Config.ReadyIngressAlgorithm))
	ejecutando := true
	for ejecutando {
		if len(ColaReady) > 0 {
			switch Config.ReadyIngressAlgorithm {
			case "FIFO":
				if len(ColaExecute) == 0 {
					firstPCB := ColaReady[0]
					nombreCPU, hayDisponible := GetCPUDisponible()
					if hayDisponible {
						ejecucion := structs.Ejecucion{
							PID: firstPCB.PID,
							PC:  firstPCB.PC,
						}

						// Marca como ejecutando
						cpu := InstanciasCPU[nombreCPU]
						cpu.Ejecutando = true
						InstanciasCPU[nombreCPU] = cpu

						// Envia el proceso
						for !PingCPU(nombreCPU) {
						}
						utils.EnviarMensaje(cpu.IP, cpu.Puerto, "dispatch", ejecucion)
						MoverPCB(firstPCB.PID, &ColaReady, &ColaExecute, structs.EstadoExec)
					}
				}
			case "SJF":
				//ejecutar SJF, no es de este checkpoint lo haremos despues (si dios quiere)
			case "SJF-SD":
				//ejecutar SJF sin desalojo, no es de este checkpoint lo haremos despues (si dios quiere)
			default:
				slog.Error(fmt.Sprintf("Algoritmo de planificacion de corto plazo no reconocido: %s", Config.ReadyIngressAlgorithm))
				ejecutando = false
			}
		}
	}
	slog.Info("Finalizando planificador de corto plazo")
}

func IniciarPlanificadores() {

	go PlanificadorCortoPlazo()
	go PlanificadorMedianoPlazo()
	bufio.NewReader(os.Stdin).ReadBytes('\n') // espera al Enter
	go PlanificadorLargoPlazo()
}

// Mueve el pcb de una lista de procesos a otra EJ: mueve de NEW a READY y cambia al nuevo estado
func MoverPCB(pid uint, origen *[]structs.PCB, destino *[]structs.PCB, estadoNuevo string) {
	for i, pcb := range *origen {
		if pcb.PID == pid {
			pcb.Estado = estadoNuevo                   // cambiar el estado del PCB
			*destino = append(*destino, pcb)           // mover a la cola destino
			*origen = slices.Delete((*origen), i, i+1) // eliminar del origen

			// Log obligatorio 3/8
			logueador.CambioDeEstado(pid, (*origen)[i].Estado, estadoNuevo)

			return
		}
	}
}

// ---------------------------- Funciones de utilidad ----------------------------//
func NuevoProceso(rutaArchInstrucciones string, tamanio int) {

	// Verifica si hay lugar disponible en memoria
	respuesta := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "check-memoria", tamanio)
	if respuesta != "OK" {
		slog.Error(fmt.Sprintf("No hay suficiente espacio en memoria. Esperando a que termine el proceso PID (%d)...", ColaExecute[0].PID))
		for ColaExecute != nil {
			// Espera a que termine el proceso ejecutando actualmente
		}
	}

	// Reserva el tamaño para memoria
	proceso := structs.NuevoProceso{
		PID:           contadorProcesos, // PID actual
		Instrucciones: rutaArchInstrucciones,
		Tamanio:       tamanio,
	}

	utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "nuevo-proceso", proceso)

	// Crea el PCB y lo inserta en NEW
	pcb := CrearPCB()
	ColaNew = append(ColaNew, pcb)
	contadorProcesos++

	// Log obligatorio 2/8
	logueador.KernelCreacionDeProceso(pcb.PID)
}

func CrearPCB() structs.PCB {
	return structs.PCB{
		PID:            contadorProcesos,
		PC:             0,
		Estado:         structs.EstadoNew,
		MetricasConteo: nil,
		MetricasTiempo: nil,
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

func PingCPU(nombre string) bool {
	instancia := InstanciasCPU[nombre]
	url := fmt.Sprintf("http://%s:%d/ping", instancia.IP, instancia.Puerto)
	_, err := http.Get(url)

	if err != nil {
		return false
	}
	return true
}

func PingIO(nombre string) bool {
	interfaz := Interfaces[nombre]
	url := fmt.Sprintf("http://%s:%d/ping", interfaz.IP, interfaz.Puerto)
	_, err := http.Get(url)

	if err != nil {
		return false
	}
	return true
}

func HandleIODisconnect(nombre string) bool {
	pid := ColaExecute[0].PID
	if !PingIO(nombre) {
		slog.Info(fmt.Sprintf("La interfaz %s fue desconectada.", nombre))
		MoverPCB(pid, &ColaExecute, &ColaExit, structs.EstadoExit)
		return true
	}
	return false
}

func LimpiarExecIO(nombre string) {
	aux := slices.Delete(ListaExecIO[nombre], 0, 1)
	ListaExecIO[nombre] = aux
}

func IniciarTimerSuspension(pid uint) {
	// Crear nuevo timer
	if _, exists := ProcesosEnTimer[pid]; exists {
		slog.Warn(fmt.Sprintf("Intento de iniciar un timer para PID %d que ya tiene uno activo. Ignorando.", pid))
		return
	}
	timer := time.NewTimer(time.Duration(Config.SuspensionTime) * time.Millisecond)
	ProcesosEnTimer[pid] = timer

	// Log de inicio del timer
	slog.Info(fmt.Sprintf("Timer de suspensión configurado para PID %d. Duración: %d ms. Será evaluado por PlanificadorMedianoPlazo.", pid, Config.SuspensionTime))
	// La lógica de expiración y movimiento ahora es manejada por PlanificadorMedianoPlazo.
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
		// Se crea una copia de los PIDs para evitar problemas si la cola es modificada por otra goroutine durante la iteración.
		var pidsEnBlockedActual []uint
		for _, pcb := range ColaBlocked { // Asumir acceso seguro o añadir mutex aquí
			pidsEnBlockedActual = append(pidsEnBlockedActual, pcb.PID)
		}

		for _, currentPid := range pidsEnBlockedActual {
			if timer, timerExists := ProcesosEnTimer[currentPid]; timerExists {
				// Verificar si el timer ha expirado de forma no bloqueante.
				select {
				case <-timer.C: // El timer ha disparado.
					slog.Info(fmt.Sprintf("PlanificadorMedianoPlazo: Timer expirado para PID %d (en ColaBlocked).", currentPid))

					respuestaMemoria := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "mover-a-swap", currentPid)
					slog.Info(fmt.Sprintf("PlanificadorMedianoPlazo: Respuesta de 'mover-a-swap' para PID %d: '%s'", currentPid, respuestaMemoria))

					MoverPCB(currentPid, &ColaBlocked, &ColaSuspBlocked, structs.EstadoWaiting)
					delete(ProcesosEnTimer, currentPid) // Eliminar el timer del mapa.
				default:
					break
				}
			}
		}
	}
}