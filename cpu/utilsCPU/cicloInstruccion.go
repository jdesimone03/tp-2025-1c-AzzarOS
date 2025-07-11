package utilsCPU

import (
	"strconv"
	"strings"
	"utils"
	"utils/logueador"
	"utils/structs"
	"sync/atomic"
)

// ---------------------------------- CICLO CPU ------------------------------------//

func Ejecucion(ctxEjecucion structs.EjecucionCPU) {
	for {
		// Decodificamos la instruccion
		instruccionCodificada := FetchAndDecode(&ctxEjecucion)
		if instruccionCodificada == nil {
			return
		}
		instruccionEjecutada := Execute(&ctxEjecucion, instruccionCodificada)
		switch instruccionEjecutada {
		case "GOTO":
			// ctxEjecucion.PC = uint(instruccionCodificada.(structs.GotoInstruction).TargetAddress)
		case "EXIT":
			return
		default:
			ctxEjecucion.PC++
		}
		if atomic.LoadInt32(&InterruptFlag) == 1 {
			// Atiende la interrupcion
			logueador.Info("PID: %d - Interrumpido, Guarda contexto en PC: %d", ctxEjecucion.PID, ctxEjecucion.PC)
			utils.EnviarMensaje(Config.IPKernel, Config.PortKernel, "guardar-contexto", ctxEjecucion)
			atomic.StoreInt32(&InterruptFlag, 0)  // Reset atómico
			return
		}

	}
}

func FetchAndDecode(ctxEjecucion *structs.EjecucionCPU) any {
	// Log obligatorio 1/11
	logueador.FetchInstruccion(ctxEjecucion.PID, ctxEjecucion.PC)
	instruccion := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "proximaInstruccion", ctxEjecucion)
	if instruccion == "PID no existe" {
		return nil
	}
	instruccionDecodificada := Decode(instruccion)
	return instruccionDecodificada
}

func Execute(ctxEjecucion *structs.EjecucionCPU, decodedInstruction any) string {
	var nombreInstruccion = utils.ParsearNombreInstruccion(decodedInstruction)

	// Por si hay una syscall
	var esSyscall bool

	switch instruccion := decodedInstruction.(type) {
	case structs.NoopInstruction:
		//hace nada
	case structs.WriteInstruction:
		Write(ctxEjecucion.PID, instruccion)
	case structs.ReadInstruction:
		Read(ctxEjecucion.PID, instruccion)
	case structs.GotoInstruction:
		ctxEjecucion.PC = uint(instruccion.TargetAddress)
	case structs.IoInstruction:
		esSyscall = true
	case structs.InitProcInstruction:
		esSyscall = true
	case structs.DumpMemoryInstruction:
		esSyscall = true
	case structs.ExitInstruction:
		esSyscall = true
	default:
		logueador.Error("llego una instruccion desconocida %v ", instruccion)
		//si llega algo inesperado
	}
	if esSyscall {
		stringPID := strconv.Itoa(int(ctxEjecucion.PID))
		stringPC := strconv.Itoa(int(ctxEjecucion.PC))
		url := "syscall/" + nombreInstruccion + "?pid=" + stringPID + "&pc=" + stringPC
		utils.EnviarMensaje(Config.IPKernel, Config.PortKernel, url, decodedInstruction)
	}
	// Log obligatorio 3/11
	logueador.InstruccionEjecutada(ctxEjecucion.PID, nombreInstruccion, decodedInstruction)
	return nombreInstruccion
}

// ---------------------------- Handlers de instrucciones ----------------------------//

// Instrucciones de memoria
func Read(pid uint, inst structs.ReadInstruction) {

	// Verificar si la página esta en Cache
	if EstaEnCache(pid, inst.Address) {
		LeerDeCache(pid, inst.Address, inst.Size)
		return
	}

	logueador.PaginaFaltanteEnCache(pid, inst.Address / ConfigMemoria.TamanioPagina) // Logueamos la pagina faltante en cache

	direccionFisica := TraducirDireccion(pid, inst.Address) // Traducimos la dirección lógica a física
	if direccionFisica == -1 {
		logueador.Info("Error al traducir la dirección lógica %d para el PID %d", inst.Address, pid)
		return
	}

	inst2 := structs.ReadInstruction{
		Address: direccionFisica, // Asignamos la dirección física
		Size:    inst.Size,       // Asignamos el tamaño a leer
		PID:     pid,             // Asignamos el PID del proceso
	}

	read := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "read", inst2)
	if read == "" {
		logueador.Info("Se leyó un string vacio")
		return
	}

	if CacheHabilitado() {
	pagina, err := PedirFrameAMemoria(pid, inst.Address, direccionFisica)
	if err != nil {
		logueador.Error("Error al pedir el frame a memoria: %v", err)
		return
	}
	logueador.Info("Página leída de memoria: %v - Agregandola a caché", pagina)
	AgregarPaginaACache(pagina)
	}

	// Log obligatorio 4/11
	logueador.LecturaMemoria(pid, inst.Address, read)
}

func Write(pid uint, inst structs.WriteInstruction) {

	// Verificar si la página esta en Cache
	if EstaEnCache(pid, inst.LogicAddress) {
		logueador.Info("Página encontrada en caché, escribiendo directamente en caché")
		EscribirEnCache(pid, inst.LogicAddress, inst.Data) // Escribimos en la caché
		return
	}
	
	logueador.PaginaFaltanteEnCache(pid, inst.LogicAddress / ConfigMemoria.TamanioPagina)

	direccionFisica := TraducirDireccion(pid, inst.LogicAddress) // Traducimos la dirección lógica a física
	if direccionFisica == -1 {
		logueador.Info("Error al traducir la dirección lógica %d para el PID %d", inst.LogicAddress, pid)
		return
	}
	logueador.Info("Direccion traducida: %d para el PID %d", direccionFisica, pid)

	inst2 := structs.WriteInstruction{
		LogicAddress: direccionFisica, // Asignamos la dirección física
		Data:         inst.Data,       // Asignamos los datos a escribir
		PID:          pid,             // Asignamos el PID del proceso
	}

	resp := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "write", inst2)
	if resp != "OK" {
		logueador.Info("Error al escribir en memoria para el PID %d, dirección %d", pid, inst.LogicAddress)
		return
	}

	if CacheHabilitado(){
		pagina, err := PedirFrameAMemoria(pid, inst.LogicAddress, direccionFisica)
		if err != nil {
			logueador.Info("Error al pedir el frame a memoria: %v", err)
			return
		}
		AgregarPaginaACache(pagina)
	}

	logueador.EscrituraMemoria(pid, inst.LogicAddress, inst.Data) // Log obligatorio 4/11
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
		logueador.Error("línea vacía")
		return nil
	}

	cmd := parts[0]
	params := parts[1:]

	instType, ok := instructionMap[cmd]
	if !ok {
		logueador.Error("comando desconocido: %s", cmd)
		return nil
	}

	switch instType {
	case structs.INST_NOOP:
		if len(params) != 0 {
			logueador.Error("NOOP no espera parámetros")
			return nil
		}
		return structs.NoopInstruction{}

	case structs.INST_WRITE:
		if len(params) != 2 {
			logueador.Error("WRITE espera 2 parámetros (Dirección, Datos)")
			return nil
		}
		addr, err := strconv.Atoi(params[0])
		if err != nil {
			logueador.Error("parámetro Dirección inválido para WRITE: %v", err)
			return nil
		}
		return structs.WriteInstruction{LogicAddress: addr, Data: params[1]}

	case structs.INST_READ:
		if len(params) != 2 {
			logueador.Error("READ espera 2 parámetros (Dirección, Tamaño)")
			return nil
		}
		addr, err := strconv.Atoi(params[0])
		if err != nil {
			logueador.Error("parámetro Dirección inválido para READ: %v", err)
			return nil
		}
		size, err := strconv.Atoi(params[1])
		if err != nil {
			logueador.Error("parámetro Tamaño inválido para READ: %v", err)
			return nil
		}
		return structs.ReadInstruction{Address: addr, Size: size}

	case structs.INST_GOTO:
		if len(params) != 1 {
			logueador.Error("GOTO espera 1 parámetro (Valor)")
			return nil
		}
		target, err := strconv.Atoi(params[0])
		if err != nil {
			logueador.Error("parámetro Valor inválido para GOTO: %v", err)
			return nil
		}
		return structs.GotoInstruction{TargetAddress: target}

	case structs.INST_IO:
		if len(params) != 2 {
			logueador.Error("IO espera 2 parámetros (Duración, Nombre)")
			return nil
		}
		duration, err := strconv.Atoi(params[1])
		if err != nil {
			logueador.Error("parámetro Duración inválido para IO: %v", err)
			return nil
		}
		return structs.IoInstruction{NombreIfaz: params[0], SuspensionTime: duration}

	case structs.INST_INIT_PROC:
		if len(params) != 2 {
			logueador.Error("INIT_PROC espera 2 parámetros (NombreProceso, TamañoMemoria)")
			return nil
		}
		memorySize, err := strconv.Atoi(params[1])
		if err != nil {
			logueador.Error("parámetro TamañoMemoria inválido para INIT_PROC: %v", err)
			return nil
		}
		return structs.InitProcInstruction{ProcessPath: params[0], MemorySize: memorySize}

	case structs.INST_DUMP_MEMORY:
		if len(params) != 0 {
			logueador.Error("DUMP_MEMORY no espera parámetros")
			return nil
		}
		return structs.DumpMemoryInstruction{}

	case structs.INST_EXIT:
		if len(params) != 0 {
			logueador.Error("EXIT no espera parámetros")
			return nil
		}
		return structs.ExitInstruction{}

	default:
		logueador.Error("parsing no implementado para: %s", cmd)
		return nil
	}
}
