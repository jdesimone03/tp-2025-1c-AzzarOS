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
	logueador.Info("Se cargara el siguiente algortimo para el planificador de largo plazo, %s", Config.SchedulerAlgorithm)
	var procesoAEnviar structs.NuevoProceso
	var firstPCB structs.PCB
	for {
		if !ColaSuspReady.Vacia() { // Toma prioridad por sobre la cola new
			firstPCB = ColaSuspReady.Obtener(0)
		} else {
			if ColaNew.Vacia() {
				continue // Espera a que haya un proceso en new
			} else {
				firstPCB = ColaNew.Obtener(0)
			}
		}

		switch Config.SchedulerAlgorithm {
		case "FIFO":
			procesoAEnviar, _ = NuevosProcesos.Obtener(firstPCB.PID)
			// Si no, no hace nada. Sigue con el bucle hasta que se libere
		case "PMCP":
			procesoMinimo, _ := NuevosProcesos.Obtener(firstPCB.PID)
			for i := range ColaNew.Longitud() {
				pcb := ColaNew.Obtener(i)
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

		// Si NO esta en los procesos en espera
		_, existe := ProcesosEnEspera.Obtener(procesoAEnviar.PID)
		if !existe {
			logueador.Info("Proceso a enviar - PID: %d, Archivo de Instrucciones: %s, Tamanio: %d", procesoAEnviar.PID, procesoAEnviar.Instrucciones, procesoAEnviar.Tamanio)
			IntentarInicializarProceso(procesoAEnviar, ColaNew)
		}

	}
}

func IntentarInicializarProceso(proceso structs.NuevoProceso, origen *structs.ColaSegura) {
	respuesta := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "check-memoria", proceso.Tamanio)
	if respuesta == "OK" {
		utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "inicializarProceso", proceso)
		MoverPCB(proceso.PID, origen, ColaReady, structs.EstadoReady)

		ProcesosEnEspera.Eliminar(proceso.PID)
	} else {
		logueador.Warn("(%d) No hay espacio en memoria para enviar el proceso. Esperando a que la memoria se libere...", proceso.PID)
		ProcesosEnEspera.Agregar(proceso.PID, proceso)
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

			for i := range ColaExecute.Longitud() {
				pcb := ColaExecute.Obtener(i)
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

	return pcbMasChico, estimadoMasChico
}
