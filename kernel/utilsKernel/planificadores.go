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
    
    for {
        // Obtener dinámicamente la cola y el PCB a procesar
        cola, firstPCB, hayProceso := ObtenerProximaColaProceso()
        
        if !hayProceso {
            continue // Espera a que haya un proceso disponible
        }

        switch Config.SchedulerAlgorithm {
        case "FIFO":
            procesoAEnviar, _ = NuevosProcesos.Obtener(firstPCB.PID)
        case "PMCP":
            procesoAEnviar = ObtenerProcesoMenorTamanio(cola)
        default:
            logueador.Error("Algoritmo de planificacion de largo plazo no reconocido: %s", Config.SchedulerAlgorithm)
            return
        }

        // Si NO esta en los procesos en espera
        _, existe := ProcesosEnEspera.Obtener(procesoAEnviar.PID)
        if !existe {
            logueador.Info("Proceso a enviar - PID: %d, Archivo de Instrucciones: %s, Tamanio: %d", procesoAEnviar.PID, procesoAEnviar.Instrucciones, procesoAEnviar.Tamanio)
            IntentarInicializarProceso(procesoAEnviar, cola)
        }
    }
}

// Función para obtener dinámicamente la próxima cola a procesar
func ObtenerProximaColaProceso() (*structs.ColaSegura, structs.PCB, bool) {
    // Prioridad 1: Cola de suspendidos listos
    if !ColaSuspReady.Vacia() {
        firstPCB := ColaSuspReady.Obtener(0)
        return ColaSuspReady, firstPCB, true
    }
    
    // Prioridad 2: Cola de nuevos procesos
    if !ColaNew.Vacia() {
        firstPCB := ColaNew.Obtener(0)
        return ColaNew, firstPCB, true
    }
    
    // No hay procesos disponibles
    return nil, structs.PCB{}, false
}

// Función para obtener el proceso con menor tamaño de una cola específica
func ObtenerProcesoMenorTamanio(cola *structs.ColaSegura) structs.NuevoProceso {
    if cola.Vacia() {
        return structs.NuevoProceso{}
    }
    
    // Inicializar con el primer proceso de la cola
    firstPCB := cola.Obtener(0)
    procesoMinimo, _ := NuevosProcesos.Obtener(firstPCB.PID)
    
    // Iterar sobre todos los procesos en la cola para encontrar el menor
    for i := 1; i < cola.Longitud(); i++ {
        pcb := cola.Obtener(i)
        nuevoProceso, _ := NuevosProcesos.Obtener(pcb.PID)
        if nuevoProceso.Tamanio < procesoMinimo.Tamanio {
            procesoMinimo = nuevoProceso
        }
    }
    
    return procesoMinimo
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
