package utilsCPU

import (
	"strconv"
	"strings"
	"utils"
	"utils/logueador"
	"utils/structs"
	"fmt"
	"net/http"
	"encoding/json"
	"math"
)

// -------------------------------- MMU --------------------------------- //

func PedirConfigMemoria() error  {
	url := fmt.Sprintf("http://%s:%d/config", Config.IPMemory, Config.PortMemory)
	logueador.Info("Solicitando configuración de Memoria en: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("falló el GET a memoria: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("memoria respondió con error HTTP %d", resp.StatusCode)
	}

	var config structs.ConfigMemoria
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return fmt.Errorf("falló el decode del JSON: %w", err)
	}

	ConfigMemoria = &config
	return nil
}

var ConfigMemoria *structs.ConfigMemoria


func nroPagina(direccionLogica int, pagesize int) int {
	return direccionLogica / pagesize
}

func desplazamiento(direccionLogica int, pagesize int) int {
	return direccionLogica % pagesize
}

func entradaNiveln(direccionlogica int, niveles int, idTabla int) int {
	pagina := direccionlogica / ConfigMemoria.TamanioPagina
	divisor := int(math.Pow(float64(ConfigMemoria.EntradasPorTabla), float64(niveles - idTabla)))
	return (pagina / divisor) % ConfigMemoria.EntradasPorTabla
}

func PedirTablaDePaginas(pid uint) *structs.Tabla {
	url := fmt.Sprintf("http://%s:%d/tabla-paginas?pid=", Config.IPMemory, Config.PortMemory) + strconv.Itoa(int(pid))
	resp, err := http.Get(url)
	if err != nil {
		logueador.Error("Error al solicitar la tabla de páginas: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logueador.Error("Error al obtener la tabla de páginas, código de estado: %d", resp.StatusCode)
		return nil
	}

	var tabla structs.Tabla
	if err := json.NewDecoder(resp.Body).Decode(&tabla); err != nil {
		logueador.Error("Error al decodificar la tabla de páginas: %v", err)
		return nil
	}

	return &tabla
}

func MMU(pid uint, direccionLogica int) int {

	desplazamiento := desplazamiento(direccionLogica, ConfigMemoria.TamanioPagina)
	tabla := PedirTablaDePaginas(pid) // Obtengo la tabla de páginas del PID

	if tabla == nil {	
		logueador.Info("No se pudo obtener la tabla de páginas para el PID %d", pid)
		return -1
	}
	
	raiz := tabla
	for nivel := 1; nivel <= ConfigMemoria.CantNiveles; nivel++ {
		entrada := entradaNiveln(direccionLogica, ConfigMemoria.CantNiveles, nivel)
		// Si llegamos al nivel final => queda buscar el frame unicamente 
		if nivel == ConfigMemoria.CantNiveles {
			if entrada >= len(raiz.Valores) || raiz.Valores[entrada] == -1 { // verifico si la entrada es válida
				logueador.Info("Dirección lógica %d no está mapeada en la tabla de páginas del PID %d", direccionLogica, pid)
				return -1 // Dirección no mapeada
			}
		frame := raiz.Valores[entrada] // Obtengo el frame correspondiente a la entrada
		return frame*ConfigMemoria.TamanioPagina + desplazamiento // Esto es el frame correspondiente a la dirección lógica 
		}
		// Si estamos en niveles intermedios => seguimos recorriendo la tabla de páginas
		if entrada >= len(raiz.Punteros) || raiz.Punteros[entrada] == nil { 
			logueador.Info("Dirección lógica %d no está mapeada en la tabla de páginas del PID %d", direccionLogica, pid)
			return -1 // Dirección no mapeada
		}
		raiz = raiz.Punteros[entrada] // Avanzamos al siguiente nivel de la tabla de páginas
	}

	logueador.Info("Error al procesar la dirección lógica %d para el PID %d", direccionLogica, pid)
	return -1 // Si llegamos hasta aca => error en el procesamiento de la dirección lógica
}

// ---------------------------------- TRADUCCIÓN DE DIRECCIONES ----------------------------------//

func TraducirDireccion(pid uint, direccion int) int {

	logueador.Info("Traduciendo dirección lógica a física")
	paginaLogica := direccion / ConfigMemoria.TamanioPagina 
	offset := desplazamiento(direccion, ConfigMemoria.TamanioPagina) // desplazamiento dentro de la página

	logueador.Info("Accediendo a TLB")
	// 1. Preguntamos a TLB
	frame := AccesoATLB(int(pid), paginaLogica) // Verificamos si la página está en la TLB
	if frame != -1{
		return frame * ConfigMemoria.TamanioPagina + offset // Retornamos la dirección física
	} 
	logueador.Info("Página no encontrada en TLB, buscando en tabla de páginas - MMU")
	// 2. Si no está en TLB, buscamos en la tabla de páginas
	direccionFisica := MMU(pid, direccion) // Obtenemos el frame físico correspondiente a la página lógica
	if direccionFisica == -1 {
		logueador.Info("Error al traducir la dirección lógica %d para el PID %d", direccion, pid)
		return -1 // Retornamos -1 para indicar que no se pudo traducir la dirección
	}

	frameFisico := direccionFisica / ConfigMemoria.TamanioPagina 

	// HUBO MISS => AGREGAR A TLB
	AgregarEntradaATLB(int(pid), paginaLogica, frameFisico) // Agregamos la entrada a la TLB
	return direccionFisica// Retornamos la dirección física
}


// ---------------------------------- CICLO CPU ------------------------------------//

func Ejecucion(ctxEjecucion structs.EjecucionCPU) {
	for {
		// Decodificamos la instruccion
		instruccionCodificada := FetchAndDecode(&ctxEjecucion)
		instruccionEjecutada := Execute(&ctxEjecucion, instruccionCodificada)
		switch instruccionEjecutada {
		case "GOTO":
			ctxEjecucion.PC = uint(instruccionCodificada.(structs.GotoInstruction).TargetAddress)
		case "EXIT":
			return
		default:
			ctxEjecucion.PC++
		}
		if InterruptFlag {
			// Atiende la interrupcion
			logueador.Info("PID: %d - Interrumpido, Guarda contexto en PC: %d", ctxEjecucion.PID, ctxEjecucion.PC)
			utils.EnviarMensaje(Config.IPKernel, Config.PortKernel, "guardar-contexto", ctxEjecucion)
			InterruptFlag = false
			return
		}

	}
}

func FetchAndDecode(ctxEjecucion *structs.EjecucionCPU) any {
	// Log obligatorio 1/11
	logueador.FetchInstruccion(ctxEjecucion.PID, ctxEjecucion.PC)
	instruccion := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "proximaInstruccion", ctxEjecucion)
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
	case structs.IOInstruction:
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
		utils.EnviarMensaje(Config.IPKernel, Config.PortKernel, "syscall/"+nombreInstruccion+"?pid="+stringPID, decodedInstruction)
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
		logueador.Info("La dirección %d ya está en la cache del PID %d", inst.Address, pid)
		LeerDeCache(pid, inst.Address, inst.Size) 
		return 
	}

	direccionFisica := TraducirDireccion(pid, inst.Address) // Traducimos la dirección lógica a física
	if direccionFisica == -1 {
		logueador.Info("Error al traducir la dirección lógica %d para el PID %d", inst.Address, pid)
		return
	}

	inst2 := structs.ReadInstruction{
		Address: direccionFisica, // Asignamos la dirección física
		Size:    inst.Size, // Asignamos el tamaño a leer
		PID:     pid, // Asignamos el PID del proceso
	}

	pagina, err := PedirFrameAMemoria(pid, inst.Address)
	if err != nil {
		logueador.Error("Error al pedir el frame a memoria: %v", err)
		return
	}

	AgregarPaginaACache(pagina)

	read := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "read", inst2)

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

	direccionFisica := TraducirDireccion(pid, inst.LogicAddress) // Traducimos la dirección lógica a física
	if direccionFisica == -1 {
		logueador.Info("Error al traducir la dirección lógica %d para el PID %d", inst.LogicAddress, pid)
		return
	}

	inst2 := structs.WriteInstruction{
		LogicAddress: direccionFisica, // Asignamos la dirección física
		Data:    inst.Data, // Asignamos los datos a escribir
		PID:     pid, // Asignamos el PID del proceso
	}

	resp := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "write" , inst2)
	if resp != "OK" {
		logueador.Info("Error al escribir en memoria para el PID %d, dirección %d", pid, inst.LogicAddress)
		return
	}

	// Si la página no estaba en cache, pedirla a memoria
	pagina, err := PedirFrameAMemoria(pid, inst.LogicAddress)
	if err != nil {
		logueador.Info("Error al pedir el frame a memoria: %v", err)
		return
	}

	logueador.Info("PAGINA CACHE CREADA: ", pagina)
	AgregarPaginaACache(pagina)
	return 
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
		duration, err := strconv.Atoi(params[0])
		if err != nil {
			logueador.Error("parámetro Duración inválido para IO: %v", err)
			return nil
		}
		return structs.IOInstruction{NombreIfaz: params[1], SuspensionTime: duration}

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
