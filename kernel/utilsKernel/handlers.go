package utilsKernel

import (
	"net/http"
	"strconv"
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
			Interfaces[interfaz.Nombre] = interfaz.Interfaz
			go PlanificadorIO(interfaz.Nombre)
			// MoverAExecIO(interfaz.Nombre)

			logueador.Info("Nueva interfaz IO: %+v", interfaz)

		case "CPU":
			instancia, err := utils.DecodificarMensaje[structs.HandshakeCPU](r)
			if err != nil {
				logueador.Error("No se pudo decodificar el mensaje (%v)", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			InstanciasCPU[instancia.Identificador] = instancia.CPU
			logueador.Info("Nueva instancia CPU: %+v", instancia)

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
			syscall, err := utils.DecodificarMensaje[structs.IOInstruction](r)
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

	TiempoEstimado[contexto.PID] = EstimarRafaga(contexto.PID)

	// Desaloja las cpu que se estén usando.
	for nombre, instancia := range InstanciasCPU {
		if instancia.Ejecutando && instancia.PID == contexto.PID {
			instancia.Ejecutando = false
			InstanciasCPU[nombre] = instancia
		}
	}

	// Busca el proceso a guardar en la cola execute
	ColaExecute.Actualizar(contexto.PID,contexto.PC)

	chCambioDeContexto <- true
	w.WriteHeader(http.StatusOK)
}

func HandleIODisconnect(w http.ResponseWriter, r *http.Request) {
	ifaz, err := utils.DecodificarMensaje[string](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje (%v)", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logueador.Warn("Se recibió notificación de desconexión de IO: %s", *ifaz)

	// Borra cualquier proceso que este ejecutando
	if ejecucion, existe := ListaExecIO.Obtener(*ifaz); existe {
		pid := ejecucion[0].PID
		Interrumpir(GetCPU(pid))
		MoverPCB(pid, ColaExecute, ColaExit, structs.EstadoExit)
		// Borro el proceso de la lista de ejecución
		ListaExecIO.EliminarPrimero(*ifaz)
	}

	if _, existe := Interfaces[*ifaz]; existe {
		delete(Interfaces, *ifaz)
	}

	w.WriteHeader(http.StatusOK)
}

func HandleIOEnd(w http.ResponseWriter, r *http.Request) {
	ifaz, err := utils.DecodificarMensaje[string](r)
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
		MoverPCB(pid, ColaBlocked, ColaReady, structs.EstadoReady)
		w.WriteHeader(http.StatusOK)
	} else {
		logueador.Error("No existe el proceso en la lista de ejecución: %s", *ifaz)
		w.WriteHeader(http.StatusBadRequest)
	}

}
