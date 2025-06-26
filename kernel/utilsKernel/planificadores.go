package utilsKernel

import (
	"bufio"
	"log/slog"
	"os"
	"time"
	"utils"
	"utils/logueador"
	"utils/structs"
)


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
			logueador.Error("Interfaz %s no encontrada, desconectando el planificador", nombre)
			return
		}
	}
}


func PlanificadorLargoPlazo() {
	logueador.Info("Se cargara el siguiente algortimo para el planificador de largo plazo, %s", Config.SchedulerAlgorithm)
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
				logueador.Error("Algoritmo de planificacion de largo plazo no reconocido: %s", Config.SchedulerAlgorithm)
				return
			}
			logueador.Info("Proceso a enviar - PID: %d, Archivo de Instrucciones: %s, Tamanio: %d", procesoAEnviar.PID, procesoAEnviar.Instrucciones, procesoAEnviar.Tamanio)
			// TODO Liberar map de nuevos procesos?
			respuesta := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "check-memoria", procesoAEnviar.Tamanio)
			if respuesta != "OK" {
				logueador.Warn("(%d) No hay espacio en memoria para enviar el proceso. Esperando a que la memoria se libere...", procesoAEnviar.PID)
				// Implementar semaforos para que espere que termine un proceso
			}
			utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "nuevo-proceso", procesoAEnviar)
			MoverPCB(procesoAEnviar.PID, ColaNew, ColaReady, structs.EstadoReady)
			//TODO timesleep?
		}
	}
}

func PlanificadorCortoPlazo() {
	logueador.Info("Se cargara el siguiente algortimo para el planificador de corto plazo, %s", Config.ReadyIngressAlgorithm)

	// var estimado = float64(Config.InitialEstimate)
	// var real int64
	var aEjecutar structs.PCB
	for {
		if ColaReady.NoVacia() {
			nombreCPU, hayDisponible := GetCPUDisponible()
			if hayDisponible {
				switch Config.ReadyIngressAlgorithm {
				case "FIFO":
					aEjecutar = ColaReady.Obtener(0)
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
					pcbMasChico := ColaReady.Obtener(0)
					for _, pcb := range(ColaReady.Cola) {
						if TiempoEstimado[pcb.PID] < TiempoEstimado[pcbMasChico.PID] {
							pcbMasChico = pcb
						}
					}
					aEjecutar = pcbMasChico
					//1 estimar todos los procesos en la cola de ready
					//2 elegir el mas chico
					//3 mandar a ejecutar el mas chico
					//4 iniciar el timer
					//5 en base al ultimo timer reestimar todos los procesos en la cola de ready
				default:
					logueador.Error("Algoritmo de planificacion de corto plazo no reconocido: %s", Config.ReadyIngressAlgorithm)
					return
				}
				ejecucion := structs.EjecucionCPU{
					PID: aEjecutar.PID,
					PC:  aEjecutar.PC,
				}

				// Marca como ejecutando
				cpu := InstanciasCPU[nombreCPU]
				cpu.Ejecutando = true
				cpu.PID = aEjecutar.PID
				InstanciasCPU[nombreCPU] = cpu

				// Envia el proceso
				TiempoEnColaExecute[aEjecutar.PID] = time.Now().UnixMilli() // Inicia el timer de ejecución, se para cuando se interrumpe
				utils.EnviarMensaje(cpu.IP, cpu.Puerto, "dispatch", ejecucion)
				MoverPCB(aEjecutar.PID, ColaReady, ColaExecute, structs.EstadoExec)
			}
		}
	}
}

func PlanificadorMedianoPlazo() {
	logueador.Info("Iniciando Planificador de Mediano Plazo.")

	for {
		slog.Debug("PlanificadorMedianoPlazo: Ejecutando ciclo de verificación de suspensión.")

		// NOTA: Para un sistema robusto, el acceso concurrente a ColaBlocked y ProcesosEnTimer
		// desde múltiples goroutines (otros planificadores, handlers) debería protegerse con mutex.
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
					logueador.Info("PlanificadorMedianoPlazo: Timer expirado para PID %d (en ColaBlocked).", currentPid)

					// Aquí se asume que el PCB con currentPid todavía está en ColaBlocked y es el que queremos mover.
					// MoverPCB buscará por PID.
					respuestaMemoria := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "mover-a-swap", currentPid)
					if respuestaMemoria != "OK" {
						logueador.Error("PlanificadorMedianoPlazo: Error al mover el PCB con PID %d a swap: '%s'", currentPid, respuestaMemoria)
						// Si no se pudo mover a swap, no se mueve el PCB y se deja en ColaBlocked.
						break // No mover el PCB, continuar con el siguiente.
					}

					logueador.Info("PlanificadorMedianoPlazo: Respuesta de 'mover-a-swap' para PID %d: '%s'", currentPid, respuestaMemoria)

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
		logueador.Info("Esperando confirmación para iniciar el planificador de largo plazo...")
		bufio.NewReader(os.Stdin).ReadBytes('\n') // espera al Enter
		go PlanificadorLargoPlazo()
	}()
}