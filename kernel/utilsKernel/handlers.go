package utilsKernel

import (
	"net/http"
	"strconv"
	"time"
	"utils"
	"utils/logueador"
	"utils/structs"
)

// ---------------------------- Handlers ----------------------------//
func HandleHandshake(tipo string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch tipo {
		case "IO":
			interfaz, err := utils.DecodificarMensaje[structs.HandshakeIO](r)
			if err != nil {
				logueador.Error("No se pudo decodificar el mensaje (%v)", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Inicializa la interfaz y el planificador
			Interfaces.Agregar(interfaz.Interfaz)
			if ListaWaitIO.NoVacia(interfaz.Interfaz.Nombre) {
				aEjecutar := ListaWaitIO.EliminarPrimero(interfaz.Interfaz.Nombre)
				DispatchIO(interfaz.Interfaz, aEjecutar)
			}

			logueador.Info("Nueva interfaz IO: %+v", interfaz)

		case "CPU":
			instancia, err := utils.DecodificarMensaje[structs.HandshakeCPU](r)
			if err != nil {
				logueador.Error("No se pudo decodificar el mensaje (%v)", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			InstanciasCPU.Agregar(instancia.CPU)
			logueador.Info("Nueva instancia CPU: %+v", instancia)

			SeñalizarCPUDisponible()

		default:
			logueador.Error("FATAL: %s no es un modulo valido.", tipo)
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func HandleSyscall(tipo string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		rawPID := r.URL.Query().Get("pid")
		pid, _ := strconv.ParseUint(rawPID, 10, 32)

		rawPC := r.URL.Query().Get("pc")
		pc, _ := strconv.ParseUint(rawPC, 10, 32)

		// Actualiza el pid y pc del proceso
		ColaExecute.Actualizar(uint(pid), uint(pc))

		// Log obligatorio 1/8
		logueador.SyscallRecibida(uint(pid), tipo)
		switch tipo {
		case "INIT_PROC":
			syscall, err := utils.DecodificarMensaje[structs.InitProcInstruction](r)
			if err != nil {
				logueador.Error("No se pudo decodificar el mensaje (%v)", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			SyscallInitProc(uint(pid), *syscall)
		case "DUMP_MEMORY":
			syscall, err := utils.DecodificarMensaje[structs.DumpMemoryInstruction](r)
			if err != nil {
				logueador.Error("No se pudo decodificar el mensaje (%v)", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			SyscallDumpMemory(uint(pid), *syscall)
		case "IO":
			syscall, err := utils.DecodificarMensaje[structs.IoInstruction](r)
			if err != nil {
				logueador.Error("No se pudo decodificar el mensaje (%v)", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			SyscallIO(uint(pid), *syscall)
		case "EXIT":
			syscall, err := utils.DecodificarMensaje[structs.ExitInstruction](r)
			if err != nil {
				logueador.Error("No se pudo decodificar el mensaje (%v)", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			SyscallExit(uint(pid), *syscall)
		default:
			logueador.Error("FATAL: %s no es un tipo de syscall valida.", tipo)
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func GuardarContexto(w http.ResponseWriter, r *http.Request) {
	contexto, err := utils.DecodificarMensaje[structs.EjecucionCPU](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje (%v)", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logueador.Info("(%d) Guardando contexto en PC: %d", contexto.PID, contexto.PC)

	nuevaRafaga := EstimarRafaga(contexto.PID)
	TiempoEstimado.Agregar(contexto.PID, nuevaRafaga)
	logueador.Debug("Reestimado PID %d (Nuevo estimado: %f)", contexto.PID, nuevaRafaga)

	// Desaloja las cpu que se estén usando.
	CPUsOcupadas.BuscarYEliminar(contexto.PID)
	//ChContextoGuardado.Señalizar(contexto.PID, struct{}{})
	SeñalizarCPUDisponible()

	// Busca el proceso a guardar en la cola execute, o en la blocked o en la susp blocked
	enExec := ColaExecute.Actualizar(contexto.PID, contexto.PC)
	if !enExec {
		enBlocked := ColaBlocked.Actualizar(contexto.PID, contexto.PC)
		if !enBlocked {
			ColaSuspBlocked.Actualizar(contexto.PID, contexto.PC)
		}
	}
	// horror

	w.WriteHeader(http.StatusOK)
}

// ---------------------------- IO ----------------------------//

func FinalizarBloqueado(pid uint) {
	_, indice := ColaBlocked.Buscar(pid)
	if indice > -1 {
		FinalizarProceso(pid, ColaBlocked)
	} else {
		FinalizarProceso(pid, ColaSuspBlocked)
	}
}

func HandleIODisconnect(w http.ResponseWriter, r *http.Request) {
	ifaz, err := utils.DecodificarMensaje[structs.InterfazIO](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje (%v)", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logueador.Warn("Se recibió notificación de desconexión de IO: %s", ifaz.Nombre)

	Interfaces.Eliminar(*ifaz)

	// Buscamos si hay otra instancia
	_, existe := Interfaces.Buscar(ifaz.Nombre)
	if !existe {
		// Se fija si pasamos a ready o susp. ready
		for ListaWaitIO.NoVacia(ifaz.Nombre) {
			exec := ListaWaitIO.EliminarPrimero(ifaz.Nombre)
			pid := exec.PID
			FinalizarBloqueado(pid)
		}
	}

	// Borra cualquier proceso que este ejecutando
	if ejecucion, existe := ListaExecIO.Obtener(*ifaz); existe {
		pid := ejecucion[0].PID

		FinalizarBloqueado(pid)
		
		// Borro el proceso de la lista de ejecución
		ListaExecIO.EliminarPrimero(*ifaz)

	}

	w.WriteHeader(http.StatusOK)
}

func HandleIOEnd(w http.ResponseWriter, r *http.Request) {
	ifaz, err := utils.DecodificarMensaje[structs.InterfazIO](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje (%v)", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ejecucion, existe := ListaExecIO.Obtener(*ifaz)
	if existe {
		pid := ejecucion[0].PID

		// Log obligatorio 5/8
		logueador.KernelFinDeIO(pid)

		// Borro el proceso de la lista de ejecución
		ListaExecIO.EliminarPrimero(*ifaz)

		// Se fija si pasamos a ready o susp. ready
		existe := MoverPCB(pid, ColaBlocked, ColaReady, structs.EstadoReady)
		if !existe { // Si no está en la cola blocked, está en la cola susp. blocked
			MoverPCB(pid, ColaSuspBlocked, ColaSuspReady, structs.EstadoSuspReady)
		}

		if ListaWaitIO.NoVacia(ifaz.Nombre) {
			aEjecutar := ListaWaitIO.EliminarPrimero(ifaz.Nombre)
			DispatchIO(*ifaz, aEjecutar)
		}

		w.WriteHeader(http.StatusOK)
	} else {
		logueador.Error("No existe el proceso en la lista de ejecución: %s", ifaz.Nombre)
		w.WriteHeader(http.StatusBadRequest)
	}

}

func EstimarRafaga(pid uint) float64 {
	estimadoAnterior, _ := TiempoEstimado.Obtener(pid)
	tiempoEnExecute, _ := TiempoEnColaExecute.Obtener(pid)
	realAnterior := time.Now().UnixMilli() - tiempoEnExecute
	return float64(realAnterior)*Config.Alpha + (1-Config.Alpha)*estimadoAnterior
}
