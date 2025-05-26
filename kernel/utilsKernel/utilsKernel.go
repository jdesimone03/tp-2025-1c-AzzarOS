package utilsKernel

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"time"
	"utils"
	"utils/config"
	"utils/structs"
)

// variables Config
var Config = config.CargarConfiguracion[config.ConfigKernel]("config.json")

// Colas de estados de los procesos
var ColaNew []structs.PCB
var ColaReady []structs.PCB
var ColaExecute []structs.PCB
var ColaBlocked []structs.PCB
var ColaExit []structs.PCB

var contadorProcesos uint = 0

// scheduler_algorithm: LARGO plazo
// ready_ingress_algorithm: CORTO plazo

var InstanciasCPU = make(map[string]structs.CPU)
var Interfaces = make(map[string]structs.Interfaz)

var ListaExecIO = make(map[string][]structs.EsperaIO)
var ListaWaitIO = make(map[string][]structs.EsperaIO)

// Handlers de endpoints
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

			slog.Info(fmt.Sprintf("Me llego la siguiente interfaz: %+v", interfaz))

		case "CPU":
			instancia, err := utils.DecodificarMensaje[structs.HandshakeCPU](r)
			if err != nil {
				slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Error al decodificar mensaje"))
				return
			}

			InstanciasCPU[instancia.Identificador] = instancia.CPU
			slog.Info(fmt.Sprintf("Me llego la siguiente cpu: %+v", instancia))

		default:
			slog.Error(fmt.Sprintf("FATAL: %s no es un modulo valido.", tipo))
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func HandleSyscall(tipo string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

// Syscalls
// No ejecuta directamente sino que lo encola en el planificador. El planificador despues tiene que ejecutarse al momento de iniciar la IO
func SyscallIO(peticion structs.IOInstruction) {
	pid := ColaExecute[0].PID
	nombre := peticion.NombreIfaz
	tiempoMs := peticion.SuspensionTime

	_, encontrada := Interfaces[nombre]
	if encontrada {
		espera := structs.EsperaIO{
			PID:      pid,
			TiempoMs: tiempoMs,
		}
		if len(ListaExecIO[nombre]) > 0 {
			// Enviar proceso a BLOCKED
			MoverPCB(pid, &ColaExecute, &ColaBlocked, structs.EstadoBlocked)

			// Enviar proceso a ListaWaitIO
			ListaWaitIO[nombre] = append(ListaWaitIO[nombre], espera)
		} else {
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

// Planificadores
func PlanificadorIO(nombre string) {
	for {
		interfaz, encontrada := Interfaces[nombre]
		if encontrada {
			lista := ListaExecIO[nombre]
			if len(lista) > 0 {
				// Enviar al IO el PID y el tiempo en ms
				proc := lista[0]

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
					aux := slices.Delete(ListaExecIO[nombre], 0, 1)
					ListaExecIO[nombre] = aux

					// Log obligatorio 5/8
					slog.Info(fmt.Sprintf("## (%d) finalizó IO y pasa a READY", proc.PID))
					MoverPCB(proc.PID, &ColaExecute, &ColaReady, structs.EstadoReady)

				case <-ctx.Done():
					// SI HAY DESCONEXION DE IO
					slog.Error(fmt.Sprintf("Timeout excedido para el proceso %d en la interfaz %s", proc.PID, nombre))
					MoverPCB(proc.PID, &ColaExecute, &ColaExit, structs.EstadoExit)
					delete(Interfaces, nombre)

					// Borro el proceso de la lista de ejecución
					aux := slices.Delete(ListaExecIO[nombre], 0, 1)
					ListaExecIO[nombre] = aux
					return // Si se desconecta el io hay que desconectar el planificador
				}
			}
			aux := ListaWaitIO[nombre]
			if len(aux) > 0 {
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
	for {
		if ColaNew != nil {
			switch Config.SchedulerAlgorithm {
			case "FIFO":
				firstPCB := ColaNew[0]
				MoverPCB(firstPCB.PID, &ColaNew, &ColaReady, structs.EstadoReady)
				// Si no, no hace nada. Sigue con el bucle hasta que se libere
			case "PMCP":
				//ejecutar PMCP, no es de este checkpoint lo haremos despues (si dios quiere)
			default:
				slog.Error(fmt.Sprintf("Algoritmo de planificacion no reconocido: %s", Config.SchedulerAlgorithm))
			}
		}
	}

}

func PlanificadorCortoPlazo() {
	if ColaReady != nil {
		switch Config.ReadyIngressAlgorithm {
		case "FIFO":
			if ColaExecute == nil {
				firstPCB := ColaReady[0]
				MoverPCB(firstPCB.PID, &ColaReady, &ColaExecute, structs.EstadoExec)
			}
		case "SJF":
			//ejecutar SJF, no es de este checkpoint lo haremos despues (si dios quiere)
		case "SJF-SD":
			//ejecutar SJF sin desalojo, no es de este checkpoint lo haremos despues (si dios quiere)
		default:
			slog.Error(fmt.Sprintf("Algoritmo de planificacion no reconocido: %s", Config.ReadyIngressAlgorithm))
		}
	}
}

/*
func init() {
	go planificadorLargoPlazo()
	go planificadorCortoPlazo()
}
*/

// Mueve el pcb de una lista de procesos a otra EJ: mueve de NEW a READY y cambia al nuevo estado
func MoverPCB(pid uint, origen *[]structs.PCB, destino *[]structs.PCB, estadoNuevo string) {
	for i, pcb := range *origen {
		if pcb.PID == pid {
			pcb.Estado = estadoNuevo // cambiar el estado del PCB
			// Log obligatorio 3/8
			slog.Info(fmt.Sprintf("## (%d) pasa del estado %s al estado %s", pid, (*origen)[i].Estado, estadoNuevo))
			*destino = append(*destino, pcb)                    // mover a la cola destino
			*origen = append((*origen)[:i], (*origen)[i+1:]...) // eliminar del origen
			return
		}
	}
}

// ---------------------------- Funciones de prueba ----------------------------//
func NuevoProceso(rutaArchInstrucciones string, tamanio int) {

	// Verifica si hay lugar disponible en memoria
	respuesta := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "check-memoria", tamanio)
	if respuesta != "OK" {
		slog.Error(fmt.Sprintf("No hay suficiente espacio en memoria. Esperando a que termine el proceso PID (%d)...",ColaExecute[0].PID))
		for(ColaExecute != nil ){
			// Espera a que termine el proceso ejecutando actualmente
		}
	}

	// Reserva el tamaño para memoria
	proceso := structs.Proceso{
		PID: contadorProcesos, // PID actual
		Instrucciones: rutaArchInstrucciones,
		Tamanio: tamanio,
	}

	utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "nuevo-proceso", proceso)

	// Crea el PCB y lo inserta en NEW
	pcb := CrearPCB()
	ColaNew = append(ColaNew, pcb)
	contadorProcesos++
	
	// Log obligatorio 2/8
	slog.Info(fmt.Sprintf("## (%d) Se crea el proceso - Estado: NEW", pcb.PID))
}
	
	/* func configurarProceso(pid uint, rutaArchInstrucciones string, tamanio int) structs.Proceso {
		return structs.Proceso{
		PID:           pid,
		Instrucciones: rutaArchInstrucciones,
		Tamanio:       tamanio,
	}
} */

func CrearPCB() structs.PCB {
	return structs.PCB{
		PID:            contadorProcesos,
		PC:             0,
		Estado:         structs.EstadoNew,
		MetricasConteo: nil,
		MetricasTiempo: nil,
	}
}

//-------------------------------------------------------------------------------//