package utilsKernel

import (
	"utils/logueador"
	"utils/structs"
)

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
		logueador.Error("La interfaz %s no existe en el sistema", nombre)

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