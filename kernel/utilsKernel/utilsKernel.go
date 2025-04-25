package utilsKernel

import (
	"fmt"
	"log/slog"
	"net/http"
	"utils"
	"utils/config"
	"utils/structs"
)

// variables Config
var Config = config.CargarConfiguracion[config.ConfigKernel]("config.json")

// Colas de estados de los procesos
var ColaNew []structs.PCB
var ColaReady []structs.PCB

// scheduler_algorithm: LARGO plazo
// ready_ingress_algorithm: CORTO plazo

var InstanciasCPU = make(map[string]structs.CPU)

var ListaExecIO = make(map[string][]*structs.PCB)
var ListaWaitIO = make(map[string][]*structs.PCB)
var Interfaces = make(map[string]structs.Interfaz)

func SyscallIO(nombre string, tiempoMs int) {
	interfaz, encontrada := Interfaces[nombre]
	if encontrada {
		if len(ListaExecIO["nombre"]) != 0 {
			// Enviar proceso a BLOCKED
			// Enviar proceso a ListaWaitIO
		} else {
			// Enviar al IO el PID y el tiempo en ms
			interfaz.Puerto++ // esto despues lo borramos, es para que guido pueda probar su funcion sin que llore el codigo
		}
		
		
		
	} else {
		slog.Error(fmt.Sprintf("La interfaz %s no existe en el sistema",nombre))
		// Enviar proceso a EXIT
		
	}
}

// Handlers de endpoints
func RecibirInterfaz(w http.ResponseWriter, r *http.Request) {
	interfaz, err := utils.DecodificarMensaje[structs.PeticionIO](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	Interfaces[interfaz.Nombre] = interfaz.Interfaz
	slog.Info(fmt.Sprintf("Me llego la siguiente interfaz: %+v", interfaz))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func RecibirCPU(w http.ResponseWriter, r *http.Request) {
	instancia, err := utils.DecodificarMensaje[structs.PeticionCPU](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	InstanciasCPU[instancia.Identificador] = instancia.CPU
	slog.Info(fmt.Sprintf("Me llego la siguiente cpu: %+v", instancia))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func HandleSyscall(w http.ResponseWriter, r *http.Request) {
	// Despues le hacemos un case switch para cada syscall diferente
	peticion, err := utils.DecodificarMensaje[structs.PeticionKernel](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}
	ifazElegida := Interfaces[peticion.NombreIfaz]
	peticion.SuspensionTime = Config.SuspensionTime
	utils.EnviarMensaje(ifazElegida.IP, ifazElegida.Puerto, "peticionKernel", peticion)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}


//NOSE NI POR QUE HICE ESTA FUNCION LA VERDAD
/*
func AdministrarColas(pcb structs.PCB) {
	switch pcb.Estado {
	case "NEW":
		utils.NuevoProceso(pcb)
		ColaNew = append(ColaNew, pcb)
	case "READY":
		//TODO: Agregar a la cola de READY
	case "EXEC":
		//TODO: Agregar a la cola de EXEC
	case "BLOCKED":
		//TODO: Agregar a la cola de BLOCKED
	case "EXIT":
		//TODO: Agregar a la cola de EXIT
	case "SUSP_BLOCKED":
		//TODO: Agregar a la cola de SUSP_BLOCKED
	case "SUSP_READY":
		//TODO: Agregar a la cola de SUSP_READY
	default:
		slog.Error(fmt.Sprintf("Estado de PCB no reconocido: %s", pcb.Estado))
	}
}*/




func PlanificadorLargoPlazo(pcb structs.PCB) {
	if ColaNew == nil {
		//SE ENVIA PEDIDO A MEMORIA, SI ES OK SE MANDA A READY
		//ASUMIMOS EL CAMINO LINDO POR QUE NO ESTA HECHO LO DE MEMORIA
		MoverPCB(pcb.PID, &ColaNew, &ColaReady, structs.EstadoReady)		
		
	}else {
		switch Config.SchedulerAlgorithm {
		case "FIFO":
			//ejecutar FIFO
		case "SJF":
			//ejecutar SJF
		default:
			slog.Error(fmt.Sprintf("Algoritmo de planificacion no reconocido: %s", Config.SchedulerAlgorithm))
		}
	}
}


//Mueve el pcb de una lista de procesos a otra EJ: mueve de NEW a READY y cambia al nuevo estado 
func MoverPCB(pid uint, origen *[]structs.PCB, destino *[]structs.PCB,estadoNuevo string) {
    for i, pcb := range *origen {
        if pcb.PID == pid {
			pcb.Estado = estadoNuevo // cambiar el estado del PCB
			slog.Info(fmt.Sprintf("## %d pasa del estado %s al estado %s", pid,(*origen)[i].Estado, estadoNuevo))
            *destino = append(*destino, pcb) // mover a la cola destino
            *origen = append((*origen)[:i], (*origen)[i+1:]...) // eliminar del origen
            return
        }
    }
}




//---------------------------- Funciones de prueba ----------------------------//
func NuevoProceso(pid uint) structs.PCB {
	var pcb = CrearPCB(pid, 0, structs.EstadoNew)
	ColaNew = append(ColaNew, pcb)
	slog.Info(fmt.Sprintf("Se agregó el proceso %d a la cola de new", pcb.PID))
	return pcb
}

func CrearPCB(pid uint,pc uint, estado string) structs.PCB {
	slog.Info(fmt.Sprintf("Se ha creado el proceso %d", pid))
	return structs.PCB{
		PID:    pid,
		PC:     pc,
		Estado: estado,
		MetricasConteo: nil,
		MetricasTiempo: nil,
	}
}
//-------------------------------------------------------------------------------//