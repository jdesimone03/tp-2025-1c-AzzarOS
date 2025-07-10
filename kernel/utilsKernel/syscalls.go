package utilsKernel

import (
	"utils"
	"utils/logueador"
	"utils/structs"
)

// ---------------------------- Syscalls ----------------------------//

// No ejecuta directamente sino que lo encola en el planificador. El planificador despues tiene que ejecutarse al momento de iniciar la IO
func SyscallIO(pid uint, instruccion structs.IoInstruction) {

	ok := BuscarEInterrumpir(pid)
	if !ok {
		logueador.Error("Error al buscar e interrumpir la cpu con el PID %d", pid)
		return
	}

	nombre := instruccion.NombreIfaz
	tiempoMs := instruccion.SuspensionTime

	ifz, existe := Interfaces.Buscar(nombre)
	if existe {

		logueador.Debug("Interfaz %s encontrada: %+v", nombre, ifz)

		ejecucion := structs.EjecucionIO{
			PID:      pid,
			TiempoMs: tiempoMs,
		}

		// Enviar proceso a BLOCKED
		MoverPCB(pid, ColaExecute, ColaBlocked, structs.EstadoBlocked)

		instancia, hayDisponible := BuscarIODisponible(nombre)

		if hayDisponible {
			logueador.Debug("Disponible %s encontrada: %+v", nombre, instancia)
			// Enviar al proceso a ejecutar el IO
			DispatchIO(instancia, ejecucion)
		} else {
			// Enviar al proceso a la lista de espera de la IO
			logueador.Debug("Envio PID %d a la lista de espera de la IO %s", pid, nombre)
			ListaWaitIO.Agregar(nombre, ejecucion)
			logueador.Debug("ListaWaitIO[%s] = %+v", nombre, ListaWaitIO.Map[nombre])
		}
	} else {
		logueador.Error("La interfaz %s no existe en el sistema", nombre)

		// Enviar proceso a EXIT
		FinalizarProceso(pid, ColaExecute)
	}
}

func SyscallDumpMemory(pid uint, instruccion structs.DumpMemoryInstruction) {

	ok := BuscarEInterrumpir(pid)
	if !ok {
		logueador.Error("Error al buscar e interrumpir la cpu con el PID %d", pid)
		return
	}

	MoverPCB(pid, ColaExecute, ColaBlocked, structs.EstadoBlocked)

	respuesta := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "memoryDump", pid)
	if respuesta != "OK" {
		logueador.Error("Error al realizar el dump de memoria: %s", respuesta)
		//InstanciasCPU.Liberar(pid)
		FinalizarProceso(pid, ColaBlocked)
		return
	}

	logueador.Info("Dump de memoria realizado correctamente para el proceso %d", pid)
	MoverPCB(pid, ColaBlocked, ColaReady, structs.EstadoReady)
}

func SyscallInitProc(pid uint, instruccion structs.InitProcInstruction) {
	instrucciones := instruccion.ProcessPath
	tamanio := instruccion.MemorySize
	NuevoProceso(instrucciones, tamanio)
}

func SyscallExit(pid uint, instruccion structs.ExitInstruction) {
	FinalizarProceso(pid, ColaExecute)
}

func VerificarInicializacion() {
	// Iterar de atrás hacia adelante para evitar problemas con índices
	for i := range ColaSuspReady.Longitud() {
		if i >= ColaSuspReady.Longitud() {
			break // Saltar si el índice ya no es válido
		}
		pcb := ColaSuspReady.Obtener(i)
		_, existe := ProcesosEnEspera.Obtener(pcb.PID)
		if existe {
			ChMemoriaLiberada.Señalizar(pcb.PID, struct{}{})
		}
	}

	if ColaSuspReady.Vacia() {
		for i := range ColaNew.Longitud() {
			if i >= ColaNew.Longitud() {
				continue // Saltar si el índice ya no es válido
			}
			pcb := ColaNew.Obtener(i)
			_, existe := ProcesosEnEspera.Obtener(pcb.PID)
			if existe {
				ChMemoriaLiberada.Señalizar(pcb.PID, struct{}{})
			}
		}
	}
}

func BuscarIODisponible(nombre string) (structs.InterfazIO, bool) {
	logueador.Debug("Interfaces disponibles: %+v", Interfaces.Cola)
	logueador.Debug("Espera de IO %s: %+v", nombre, ListaWaitIO.Map[nombre])

	for i := range Interfaces.Longitud() {
		ifaz := Interfaces.Obtener(i)

		if ifaz.Nombre == nombre {
			mxBusquedaIO.Lock()

			logueador.Debug("Exec de IO %s: %+v", nombre, ListaExecIO.Map[ifaz])

			lista, _ := ListaExecIO.Obtener(ifaz)
			if len(lista) == 0 {
				mxBusquedaIO.Unlock()
				return ifaz, true
			}

			mxBusquedaIO.Unlock()
		}
	}
	logueador.Debug("La IO %s no esta disponible", nombre)
	return structs.InterfazIO{}, false
}
