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

var ColaNew []structs.PCB
var ColaReady []structs.PCB

// scheduler_algorithm: LARGO plazo
// ready_ingress_algorithm: CORTO plazo

var ListaExecIO = make(map[string][]*structs.PCB)
var ListaWaitIO = make(map[string][]*structs.PCB)
var Interfaces = make(map[string]structs.Interfaz)

//func syscallIO(nombre string, tiempoMs int) {
//	interfaz, encontrada := Interfaces[nombre]
//	if encontrada {
//		
//		// Enviar proceso a BLOCKED
//		
//		// Agregar proceso a la cola de bloqueados por la IO solicitada
//	} else {
//		slog.Error(fmt.Sprintf("La interfaz %s no existe en el sistema",nombre))
//		// Enviar proceso a EXIT
//		
//	}
//}

// Handlers de endpoints
func RecibirInterfaz(w http.ResponseWriter, r *http.Request) {
	interfaz, err := utils.DecodificarMensaje[structs.PeticionIO](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	addInterfaz(interfaz.Nombre, interfaz.Interfaz)
	slog.Info(fmt.Sprintf("Me llego la siguiente interfaz: %+v", interfaz))

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

func addInterfaz(nombre string, ifaz structs.Interfaz) {
	Interfaces[nombre] = ifaz
}

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
}

func PlanificadorLargoPlazo(pcb structs.PCB) {
	if ColaNew == nil {
		//SE ENVIA PEDIDO A MEMORIA, SI ES OK SE MANDA A READY
		//ASUMIMOS EL CAMINO LINDO POR QUE NO ESTA HECHO LO DE MEMORIA
		pcb.Estado = structs.EstadoReady
		MoverPCB(pcb.PID, &ColaNew, &ColaReady)		
		
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

func MoverPCB(pid uint, origen *[]structs.PCB, destino *[]structs.PCB) {
    for i, pcb := range *origen {
        if pcb.PID == pid {
            *destino = append(*destino, pcb) // mover a la cola destino
            *origen = append((*origen)[:i], (*origen)[i+1:]...) // eliminar del origen
            return
        }
    }
}




//---------------------------- Funciones de prueba ----------------------------//
func NuevoProceso() structs.PCB {
	var pcb = CrearPCB(1, 0, structs.EstadoNew)
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