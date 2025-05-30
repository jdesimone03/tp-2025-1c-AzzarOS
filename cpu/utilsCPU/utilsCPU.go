package utilsCPU

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"utils"
	"utils/config"
	"utils/logueador"
	"utils/structs"
)

var Config config.ConfigCPU
var Ejecutando structs.EjecucionCPU
var InterruptFlag = make(map[uint]bool)

func PingCPU(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func RecibirInterrupcion(w http.ResponseWriter, r *http.Request) {
	pid, err := utils.DecodificarMensaje[uint](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Log obligatorio 2/11
	logueador.InterrupcionRecibida()
	InterruptFlag[*pid] = true

	w.WriteHeader(http.StatusOK)
}

func RecibirEjecucion(w http.ResponseWriter, r *http.Request) {
	ejecucion, err := utils.DecodificarMensaje[structs.EjecucionCPU](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Solicitar a memoria la siguiente instruccion para la ejecución
	// InstruccionCodificada = FETCH(PCB.ProgramCounter)
	Ejecutando = *ejecucion

	go Ejecucion()

	w.WriteHeader(http.StatusOK)
}

func Ejecucion() {
	for {
		// Decodificamos la instruccion
		instruccionCodificada, terminoEjecucion := FetchAndDecode(Ejecutando)
		if terminoEjecucion {
			return
		}
		Ejecutando.PC++
		Execute(instruccionCodificada)
		if InterruptFlag[Ejecutando.PID] {
			// Atiende la interrupcion
			slog.Info(fmt.Sprintf("PID: %d - Interrumpido, Guarda contexto en PC: %d", Ejecutando.PID, Ejecutando.PC))
			utils.EnviarMensaje(Config.IPKernel, Config.PortKernel, "guardar-contexto", Ejecutando)
			InterruptFlag[Ejecutando.PID] = false
			return
		}
	}
}

func FetchAndDecode(peticion structs.EjecucionCPU) (any, bool) {
	// Log obligatorio 1/11
	logueador.FetchInstruccion(peticion.PID, peticion.PC)

	instruccion := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "fetch", peticion)
	if instruccion == "" {
		return nil, true
	}
	instruccionDecodificada := Decode(instruccion)
	return instruccionDecodificada, false
}

//AUMENTAR PC NO SE EN QUE MOMENTO SE HACE QUIERO VER LA TEORIA ANTES DE IMPLEMENTARLO

func Execute(decodedInstruction any) {
	var nombreInstruccion string
	switch instruccion := decodedInstruction.(type) {
	case structs.NoopInstruction:
		nombreInstruccion = "NOOP"
		//hace nada
	case structs.WriteInstruction:
		nombreInstruccion = "WRITE"
		//hace lo que tenga quer hacer
	case structs.ReadInstruction:
		nombreInstruccion = "READ"
		//hace lo que tenga quer hacer
	case structs.GotoInstruction:
		nombreInstruccion = "GOTO"
		//hace lo que tenga quer hacer
		Ejecutando.PC = uint(instruccion.TargetAddress)
	case structs.IOInstruction:
		nombreInstruccion = "IO"
		utils.EnviarMensaje(Config.IPKernel, Config.PortKernel, "syscall/IO", instruccion)
	case structs.InitProcInstruction:
		nombreInstruccion = "INIT_PROC"
		utils.EnviarMensaje(Config.IPKernel, Config.PortKernel, "syscall/INIT_PROC", instruccion)
	case structs.DumpMemoryInstruction:
		nombreInstruccion = "DUMP_MEMORY"
		utils.EnviarMensaje(Config.IPKernel, Config.PortKernel, "syscall/DUMP_MEMORY", instruccion)
	case structs.ExitInstruction:
		nombreInstruccion = "EXIT"
		utils.EnviarMensaje(Config.IPKernel, Config.PortKernel, "syscall/EXIT", instruccion)
	default:
		slog.Error(fmt.Sprintf("llego una instruccion desconocida %v ", instruccion))
		//si llega algo inesperado
	}
	// Log obligatorio 3/11
	logueador.InstruccionEjecutada(Ejecutando.PID, nombreInstruccion, decodedInstruction)
}

// PROPUESTA FUNCION PARSEO DE COMANDOS

// Mapa para pasar de  string a InstructionType (nos serviria para el parsing)
var instructionMap = map[string]structs.InstructionType{
	"NOOP":        structs.INST_NOOP,
	"WRITE":       structs.INST_WRITE,
	"READ":        structs.INST_READ,
	"GOTO":        structs.INST_GOTO,
	"IO":          structs.INST_IO,
	"INIT_PROC":   structs.INST_INIT_PROC,
	"DUMP_MEMORY": structs.INST_DUMP_MEMORY,
	"EXIT":        structs.INST_EXIT,
}

func Decode(line string) any {
	parts := strings.Fields(line) // Divide por espacios
	if len(parts) == 0 {
		slog.Error("línea vacía")
		return nil
	}

	cmd := parts[0]
	params := parts[1:]

	instType, ok := instructionMap[cmd]
	if !ok {
		slog.Error(fmt.Sprintf("comando desconocido: %s", cmd))
		return nil
	}

	switch instType {
	case structs.INST_NOOP:
		if len(params) != 0 {
			slog.Error("NOOP no espera parámetros")
			return nil
		}
		return structs.NoopInstruction{}

	case structs.INST_WRITE:
		if len(params) != 2 {
			slog.Error("WRITE espera 2 parámetros (Dirección, Datos)")
			return nil
		}
		addr, err := strconv.Atoi(params[0])
		if err != nil {
			slog.Error(fmt.Sprintf("parámetro Dirección inválido para WRITE: %v", err))
			return nil
		}
		return structs.WriteInstruction{Address: addr, Data: params[1]}

	case structs.INST_READ:
		if len(params) != 2 {
			slog.Error("READ espera 2 parámetros (Dirección, Tamaño)")
			return nil
		}
		addr, err := strconv.Atoi(params[0])
		if err != nil {
			slog.Error(fmt.Sprintf("parámetro Dirección inválido para READ: %v", err))
			return nil
		}
		size, err := strconv.Atoi(params[1])
		if err != nil {
			slog.Error(fmt.Sprintf("parámetro Tamaño inválido para READ: %v", err))
			return nil
		}
		return structs.ReadInstruction{Address: addr, Size: size}

	case structs.INST_GOTO:
		if len(params) != 1 {
			slog.Error("GOTO espera 1 parámetro (Valor)")
			return nil
		}
		target, err := strconv.Atoi(params[0])
		if err != nil {
			slog.Error(fmt.Sprintf("parámetro Valor inválido para GOTO: %v", err))
			return nil
		}
		return structs.GotoInstruction{TargetAddress: target}

	case structs.INST_IO:
		if len(params) != 2 {
			slog.Error("IO espera 2 parámetros (Duración, Nombre)")
			return nil
		}
		duration, err := strconv.Atoi(params[0])
		if err != nil {
			slog.Error(fmt.Sprintf("parámetro Duración inválido para IO: %v", err))
			return nil
		}
		return structs.IOInstruction{NombreIfaz: params[1], SuspensionTime: duration}

	case structs.INST_INIT_PROC:
		if len(params) != 2 {
			slog.Error("INIT_PROC espera 2 parámetros (NombreProceso, TamañoMemoria)")
			return nil
		}
		memorySize, err := strconv.Atoi(params[1])
		if err != nil {
			slog.Error(fmt.Sprintf("parámetro TamañoMemoria inválido para INIT_PROC: %v", err))
			return nil
		}
		return structs.InitProcInstruction{ProcessPath: params[0], MemorySize: memorySize}

	case structs.INST_DUMP_MEMORY:
		if len(params) != 0 {
			slog.Error("DUMP_MEMORY no espera parámetros")
			return nil
		}
		return structs.DumpMemoryInstruction{}

	case structs.INST_EXIT:
		if len(params) != 0 {
			slog.Error("EXIT no espera parámetros")
			return nil
		}
		return structs.ExitInstruction{}

	default:
		slog.Error(fmt.Sprintf("parsing no implementado para: %s", cmd))
		return nil
	}
}
