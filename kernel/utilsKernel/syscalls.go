package utilsKernel

import (
	"utils"
	"utils/logueador"
	"utils/structs"
)

// ---------------------------- Syscalls ----------------------------//
// No ejecuta directamente sino que lo encola en el planificador. El planificador despues tiene que ejecutarse al momento de iniciar la IO
func SyscallIO(pid uint, instruccion structs.IoInstruction) {

	Interrumpir(InstanciasCPU.BuscarCPUPorPID(pid))
	
	nombre := instruccion.NombreIfaz
	tiempoMs := instruccion.SuspensionTime

	_, encontrada := Interfaces.Obtener(nombre)
	if encontrada {
		ejecucion := structs.EjecucionIO{
			PID:      pid,
			TiempoMs: tiempoMs,
		}
		// Enviar proceso a BLOCKED
		MoverPCB(pid, ColaExecute, ColaBlocked, structs.EstadoBlocked)

		lista, _ := ListaExecIO.Obtener(nombre)
		if len(lista) > 0 {
			// Enviar proceso a ListaWaitIO
			ListaWaitIO.Agregar(nombre, ejecucion)
		} else {
			// Enviar al proceso a ejecutar el IO
			DispatchIO(nombre, ejecucion)
		}
	} else {
		logueador.Error("La interfaz %s no existe en el sistema", nombre)

		// Enviar proceso a EXIT
		MoverPCB(pid, ColaExecute, ColaExit, structs.EstadoExit)
	}
}

func SyscallDumpMemory(pid uint, instruccion structs.DumpMemoryInstruction) {

	Interrumpir(InstanciasCPU.BuscarCPUPorPID(pid))

	MoverPCB(pid, ColaExecute, ColaBlocked, structs.EstadoBlocked)

	respuesta := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "memoryDump", pid)
	if respuesta != "OK" {
		logueador.Error("Error al realizar el dump de memoria: %s", respuesta)
		InstanciasCPU.Liberar(pid)
		MoverPCB(pid, ColaBlocked, ColaExit, structs.EstadoExit)
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
	respuesta := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "finalizarProceso", pid)
	if respuesta != "OK" {
		logueador.Error("Error al finalizar el proceso %d: %s", pid, respuesta)
		return
	}
	InstanciasCPU.Liberar(pid)
	MoverPCB(pid, ColaExecute, ColaExit, structs.EstadoExit) // asumimos que liberar el pcb es moverlo a exit
	//VerificarProcesos()
}

func VerificarProcesos() {
	for i := range ColaSuspReady.Longitud() {
		pcb := ColaSuspReady.Obtener(i)
		MoverPCB(pcb.PID, ColaSuspReady, ColaReady, structs.EstadoReady)
	}
	for i := range ColaNew.Longitud() {
		pcb := ColaNew.Obtener(i)
		MoverPCB(pcb.PID, ColaNew, ColaReady, structs.EstadoReady)
	}
}
