package utilsKernel

import (
	"bufio"
	"os"
	"time"
	"utils"
	"utils/logueador"
	"utils/structs"
)

func PlanificadorLargoYMedianoPlazo() {
	logueador.Info("Se cargara el siguiente algortimo para el planificador de largo plazo, %s", Config.ReadyIngressAlgorithm)

	var procesoAEnviar structs.NuevoProceso
	for {
		// Debug: mostrar cantidad en el canal antes de leer
        logueador.Debug("Canal ChColasLargoMedioPlazo tiene %d elementos", len(ChColasLargoMedioPlazo))
        
        // Esperar hasta que haya procesos disponibles
        <-ChColasLargoMedioPlazo
        
        logueador.Debug("Canal ChColasLargoMedioPlazo después de leer tiene %d elementos", len(ChColasLargoMedioPlazo))

		// Obtener dinámicamente la cola y el PCB a procesar
		cola, firstPCB, hayProceso := ObtenerProximaColaProceso()

		if hayProceso {
			switch Config.ReadyIngressAlgorithm {
			case "FIFO":
				procesoAEnviar, _ = NuevosProcesos.Obtener(firstPCB.PID)
			case "PMCP":
				procesoAEnviar, _ = ObtenerProcesoMenorTamanio(firstPCB.PID, cola)
			default:
				logueador.Error("Algoritmo de planificacion de largo plazo no reconocido: %s", Config.SchedulerAlgorithm)
				return
			}

			IntentarInicializarProceso(procesoAEnviar, cola)

		}

		// Verificar si aún hay procesos para procesar
		_, _, hayMasProcesos := ObtenerProximaColaProceso()
		if hayMasProcesos {
			SeñalizarProcesoEnLargoMedioPlazo()
		}
	}
}

func PlanificadorCortoPlazo() {
	logueador.Info("Se cargara el siguiente algortimo para el planificador de corto plazo, %s", Config.SchedulerAlgorithm)

	var aEjecutar structs.PCB
	encontroVictima := false
	for {
		<-ChColaReady

		for !ColaReady.Vacia() {
			<-ChCPUDisponible
			cpu, hayDisponible := BuscarCPUDisponible()

			if !hayDisponible {
				if Config.SchedulerAlgorithm == "SRT" {
					var estimadoMasChico float64
					aEjecutar, estimadoMasChico = ObtenerMasChico()

					encontroVictima = false

					for i := range ColaExecute.Longitud() {
						pcb := ColaExecute.Obtener(i)
						estimadoActual, _ := TiempoEstimado.Obtener(pcb.PID)
						if estimadoMasChico < estimadoActual {
							// Manda a ejecutar el mas chico
							ok := BuscarEInterrumpir(pcb.PID)
							if ok {
								<-ChCPUDisponible
								MoverPCB(pcb.PID, ColaExecute, ColaReady, structs.EstadoReady)
								logueador.DesalojoSRT(pcb.PID)
								cpu, hayDisponible = BuscarCPUDisponible()
								encontroVictima = true
								break
							}
						}
					}
					if !encontroVictima {
						break // Salir del bucle interno
					}
				} else {
					break
				}
			}

			if hayDisponible {
				switch Config.SchedulerAlgorithm {
				case "FIFO":
					aEjecutar = ColaReady.Obtener(0)
				case "SJF":
					aEjecutar, _ = ObtenerMasChico()
				case "SRT":
					if !encontroVictima {
						aEjecutar, _ = ObtenerMasChico()
					}
				default:
					logueador.Error("Algoritmo de planificacion de corto plazo no reconocido: %s", Config.ReadyIngressAlgorithm)
					return
				}
				logueador.Info("Por algoritmo %s se eligió al proceso %d", Config.ReadyIngressAlgorithm, aEjecutar.PID)

				ejecucion := structs.EjecucionCPU{
					PID: aEjecutar.PID,
					PC:  aEjecutar.PC,
				}

				// Marca como ejecutando
				CPUsOcupadas.Agregar(cpu.Nombre, ejecucion)
				logueador.Info("Se ocupa la CPU %s con el proceso PID %d", cpu.Nombre, aEjecutar.PID)

				// Envia el proceso
				TiempoEnColaExecute.Agregar(aEjecutar.PID, time.Now().UnixMilli()) // Inicia el timer de ejecución, se para cuando se interrumpe
				MoverPCB(aEjecutar.PID, ColaReady, ColaExecute, structs.EstadoExec)
				utils.EnviarMensaje(cpu.IP, cpu.Puerto, "dispatch", ejecucion)
			}

			if hayDisponible {
				SeñalizarCPUDisponible()
			}
		}

		if !ColaReady.Vacia() {
			SeñalizarProcesoEnCortoPlazo()
		}
	}
}

func IniciarPlanificadores() {
	go PlanificadorCortoPlazo()
	go func() {
		logueador.Info("Esperando confirmación para iniciar el planificador de largo plazo...")
		bufio.NewReader(os.Stdin).ReadBytes('\n') // espera al Enter
		go PlanificadorLargoYMedianoPlazo()
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

	for i := range ColaReady.Longitud() {
		pcb := ColaReady.Obtener(i)
		estimadoActual, _ := TiempoEstimado.Obtener(pcb.PID)
		if estimadoActual < estimadoMasChico {
			pcbMasChico = pcb
			estimadoMasChico, _ = TiempoEstimado.Obtener(pcbMasChico.PID)
		}
	}

	logueador.Debug("Proceso mas chico: %d, Tiempo estimado: %f", pcbMasChico.PID, estimadoMasChico)

	return pcbMasChico, estimadoMasChico
}

// Función para obtener dinámicamente la próxima cola a procesar
func ObtenerProximaColaProceso() (*structs.ColaSegura, structs.PCB, bool) {
	// Prioridad 1: Cola de suspendidos listos
	for i := range ColaSuspReady.Longitud() {
		firstPCB := ColaSuspReady.Obtener(i)
		_, existe := ProcesosEnEspera.Obtener(firstPCB.PID)
		if !existe {
			return ColaSuspReady, firstPCB, true
		}
	}

	// Prioridad 2: Cola de nuevos procesos
	for i := range ColaNew.Longitud() {
		firstPCB := ColaNew.Obtener(i)
		_, existe := ProcesosEnEspera.Obtener(firstPCB.PID)
		if !existe {
			return ColaNew, firstPCB, true
		}
	}

	// No hay procesos disponibles
	return nil, structs.PCB{}, false
}

// Función para obtener el proceso con menor tamaño de una cola específica
func ObtenerProcesoMenorTamanio(firstPID uint, cola *structs.ColaSegura) (structs.NuevoProceso, bool) {
	if cola.Vacia() {
		return structs.NuevoProceso{}, false
	}

	// Inicializar con el primer proceso de la cola
	procesoMinimo, _ := NuevosProcesos.Obtener(firstPID)

	// Iterar sobre todos los procesos en la cola para encontrar el menor
	for i := range cola.Longitud() {
		pcb := cola.Obtener(i)
		_, enEspera := ProcesosEnEspera.Obtener(pcb.PID)
		if enEspera {
			continue
		}
		nuevoProceso, existe := NuevosProcesos.Obtener(pcb.PID)
		if existe && nuevoProceso.Tamanio < procesoMinimo.Tamanio {
			procesoMinimo = nuevoProceso
		}
	}

	return procesoMinimo, true
}

func BuscarCPUDisponible() (structs.InstanciaCPU, bool) {
	for i := range InstanciasCPU.Longitud() {
		cpu := InstanciasCPU.Obtener(i)

		mxBusquedaCPU.Lock()

		_, existe := CPUsOcupadas.Obtener(cpu.Nombre)
		if !existe {
			logueador.Debug("CPU disponible: %s", cpu.Nombre)
			mxBusquedaCPU.Unlock()
			return cpu, true
		}

		mxBusquedaCPU.Unlock()
	}
	return structs.InstanciaCPU{}, false
}
