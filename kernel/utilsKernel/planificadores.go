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
		interfaz, encontrada := Interfaces.Obtener(nombre)
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
			logueador.Error("Interfaz %s no encontrada, finalizando el planificador", nombre)
			return
		}
	}
}

func PlanificadorLargoPlazo() {
	logueador.Info("Se cargara el siguiente algortimo para el planificador de largo plazo, %s", Config.SchedulerAlgorithm)
	var procesoAEnviar structs.NuevoProceso
	for {
		if ColaNew.Vacia() {
			continue // Espera a que haya un proceso en new
		}
		firstPCB := ColaNew.Obtener(0)
		switch Config.SchedulerAlgorithm {
		case "FIFO":
			procesoAEnviar, _ = NuevosProcesos.Obtener(firstPCB.PID)
			// Si no, no hace nada. Sigue con el bucle hasta que se libere
		case "PMCP":
			procesoMinimo, _ := NuevosProcesos.Obtener(firstPCB.PID)
			for _, pcb := range ColaNew.Cola {
				nuevoProceso, _ := NuevosProcesos.Obtener(pcb.PID)
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

func PlanificadorCortoPlazo() {
	logueador.Info("Se cargara el siguiente algortimo para el planificador de corto plazo, %s", Config.ReadyIngressAlgorithm)

	// var estimado = float64(Config.InitialEstimate)
	// var real int64
	var aEjecutar structs.PCB
	for {
		if ColaReady.Vacia() {
			continue
		}

		nombreCPU, hayDisponible := InstanciasCPU.BuscarCPUDisponible()

		if !hayDisponible && Config.ReadyIngressAlgorithm == "SJF" {
			var estimadoMasChico float64
			aEjecutar, estimadoMasChico = ObtenerMasChico()

			for _, pcb := range ColaExecute.Cola {
				estimadoActual, _ := TiempoEstimado.Obtener(pcb.PID)
				if estimadoMasChico < estimadoActual {
					// Manda a ejecutar el mas chico
					cpuADesalojar := InstanciasCPU.BuscarCPUPorPID(pcb.PID)
					Interrumpir(cpuADesalojar)
					nombreCPU, hayDisponible = InstanciasCPU.BuscarCPUDisponible()
					break
				}
			}
		}

		if hayDisponible {
			switch Config.ReadyIngressAlgorithm {
			case "FIFO":
				aEjecutar = ColaReady.Obtener(0)
			case "SJF-SD":
				aEjecutar, _ = ObtenerMasChico()
			default:
				logueador.Error("Algoritmo de planificacion de corto plazo no reconocido: %s", Config.ReadyIngressAlgorithm)
				return
			}

			ejecucion := structs.EjecucionCPU{
				PID: aEjecutar.PID,
				PC:  aEjecutar.PC,
			}

			// Marca como ejecutando
			cpu := InstanciasCPU.Ocupar(nombreCPU, aEjecutar.PID)

			// Envia el proceso
			TiempoEnColaExecute.Agregar(aEjecutar.PID, time.Now().UnixMilli()) // Inicia el timer de ejecución, se para cuando se interrumpe
			MoverPCB(aEjecutar.PID, ColaReady, ColaExecute, structs.EstadoExec)
			utils.EnviarMensaje(cpu.IP, cpu.Puerto, "dispatch", ejecucion)
		}
	}
}

func PlanificadorMedianoPlazo() {
	logueador.Info("Iniciando Planificador de Mediano Plazo.")

	for {
		if ColaBlocked.Vacia() {
			slog.Debug("PlanificadorMedianoPlazo: Ejecutando ciclo de verificación de suspensión.")
			continue
		}

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

// ---------------------------- Funciones de utilidad ----------------------------//
func ObtenerMasChico() (structs.PCB, float64) {
	//1 estimar todos los procesos en la cola de ready
	//2 elegir el mas chico
	//3 mandar a ejecutar el mas chico
	//4 iniciar el timer
	//5 en base al ultimo timer reestimar todos los procesos en la cola de ready
	pcbMasChico := ColaReady.Obtener(0)
	estimadoMasChico, _ := TiempoEstimado.Obtener(pcbMasChico.PID)

	for _, pcb := range ColaReady.Cola {
		estimadoActual, _ := TiempoEstimado.Obtener(pcb.PID)
		if estimadoActual < estimadoMasChico {
			pcbMasChico = pcb
			estimadoMasChico, _ = TiempoEstimado.Obtener(pcbMasChico.PID)
		}
	}

	return pcbMasChico, estimadoMasChico
}
