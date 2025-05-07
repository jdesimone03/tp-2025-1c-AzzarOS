package utilsKernel

import (
	"fmt"
	"log/slog"
	"net/http"
	"utils"
	"utils/config"
	"utils/structs"
	"slices"
)

// variables Config
var Config = config.CargarConfiguracion[config.ConfigKernel]("config.json")

// Colas de estados de los procesos
var ColaNew []structs.PCB
var ColaReady []structs.PCB
var ColaExecute []structs.PCB
var ColaBlocked []structs.PCB
var ColaExit []structs.PCB

var contadorProcesos uint = 0

// scheduler_algorithm: LARGO plazo
// ready_ingress_algorithm: CORTO plazo

var InstanciasCPU = make(map[string]structs.CPU)

var ListaExecIO = make(map[string][]structs.EsperaIO) // nombre de io: PID
var ListaWaitIO = make(map[string][]structs.EsperaIO)
var Interfaces = make(map[string]structs.Interfaz)

// Handlers de endpoints
func HandleHandshake(tipo string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch tipo {
		case "IO":
			interfaz, err := utils.DecodificarMensaje[structs.HandshakeIO](r)
			if err != nil {
				slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Error al decodificar mensaje"))
				return
			}

			// Inicializa la interfaz y el planificador
			Interfaces[interfaz.Nombre] = interfaz.Interfaz
			go PlanificadorIO(interfaz.Nombre)
			// MoverAExecIO(interfaz.Nombre)

			slog.Info(fmt.Sprintf("Me llego la siguiente interfaz: %+v", interfaz))

		case "CPU":
			instancia, err := utils.DecodificarMensaje[structs.HandshakeCPU](r)
			if err != nil {
				slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Error al decodificar mensaje"))
				return
			}

			InstanciasCPU[instancia.Identificador] = instancia.CPU
			slog.Info(fmt.Sprintf("Me llego la siguiente cpu: %+v", instancia))

		default:
			slog.Error(fmt.Sprintf("FATAL: %s no es un modulo valido.", tipo))
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func HandleSyscall(tipo string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch tipo {
		case "IO":
			err := SyscallIO(r, w)
			if err {
				return
			}
		default:
			slog.Error(fmt.Sprintf("FATAL: %s no es un tipo de syscall valida.", tipo))
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)

	}
	// Despues le hacemos un case switch para cada syscall diferente
}

// No ejecuta directamente sino que lo encola en el planificador. El planificador despues tiene que ejecutarse al momento de iniciar la IO
func SyscallIO(r *http.Request, w http.ResponseWriter) bool {
	peticion, err := utils.DecodificarMensaje[structs.PeticionIO](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return true
	}

	pid := peticion.PID
	nombre := peticion.NombreIfaz
	tiempoMs := peticion.SuspensionTime

	_, encontrada := Interfaces[nombre]
	if encontrada {
		espera := structs.EsperaIO{
			PID:      pid,
			TiempoMs: tiempoMs,
		}
		_, isExec := ListaExecIO[nombre]
		if isExec {
			// Enviar proceso a BLOCKED
			MoverPCB(pid, &ColaExecute, &ColaBlocked, structs.EstadoBlocked)

			// Enviar proceso a ListaWaitIO
			ListaWaitIO[nombre] = append(ListaWaitIO[nombre], espera)
		} else {
			// Enviar al proceso a ejecutar el IO
			ListaExecIO[nombre] = append(ListaExecIO[nombre], espera)
		}
	} else {
		slog.Error(fmt.Sprintf("La interfaz %s no existe en el sistema", nombre))

		// Enviar proceso a EXIT
		MoverPCB(pid, &ColaExecute, &ColaExit, structs.EstadoBlocked)
	}
	return false
}

func PlanificadorIO(nombre string) {
	for {
		interfaz, encontrada := Interfaces[nombre]
		if encontrada {
			lista, hayExec := ListaExecIO[nombre]
			if hayExec {
				// Enviar al IO el PID y el tiempo en ms
				proc := lista[0]
				peticion := structs.PeticionIO{
					PID:            proc.PID,
					NombreIfaz:     nombre,
					SuspensionTime: proc.TiempoMs,
				}
				utils.EnviarMensaje(interfaz.IP, interfaz.Puerto, "peticionIO", peticion)
				// Borro el proceso de la lista de ejecucion
				aux := slices.Delete(ListaExecIO[nombre], 0, 1 )
				ListaExecIO[nombre] = aux
			}
			aux, hayEsperando := ListaWaitIO[nombre]
			if hayEsperando {
				// Borra el primer elemento en la lista de espera
				aEjecutar := aux[0]
				aux := slices.Delete(aux, 0, 1)
				ListaWaitIO[nombre] = aux
				ListaExecIO[nombre] = append(ListaExecIO[nombre], aEjecutar)
			}
		} else {
			// Si se llega a desconectar el IO, se desconecta el planificador
			// Tengo que ver como hacer para que se borre la interfaz de la lista de interfaces al momento que se desconecta
			break
		}
	}
}

func PlanificadorLargoPlazo(pcb structs.PCB) {
	if ColaNew == nil { // usar chanels, ni idea que onda eso pero me lo dijo el ayudante
		//SE ENVIA PEDIDO A MEMORIA, SI ES OK SE MANDA A READY
		//ASUMIMOS EL CAMINO LINDO POR QUE NO ESTA HECHO LO DE MEMORIA
		MoverPCB(pcb.PID, &ColaNew, &ColaReady, structs.EstadoReady)
	} else {
		switch Config.SchedulerAlgorithm {
		case "FIFO":
			FIFO()
		case "PMCP":
			//ejecutar PMCP, no es de este checkpoint lo haremos despues (si dios quiere)
		default:
			slog.Error(fmt.Sprintf("Algoritmo de planificacion no reconocido: %s", Config.SchedulerAlgorithm))
		}
	}
}

func FIFO() {
	if ColaNew != nil {
		firstPCB := ColaNew[0]
		MoverPCB(firstPCB.PID, &ColaNew, &ColaReady, structs.EstadoReady)
	}
}

func PlanificadorCortoPlazo() {
	if ColaReady != nil {
		switch Config.ReadyIngressAlgorithm {
		case "FIFO":
			FIFO()
		case "SJF":
			//ejecutar SJF, no es de este checkpoint lo haremos despues (si dios quiere)
		case "SJF-SD":
			//ejecutar SJF sin desalojo, no es de este checkpoint lo haremos despues (si dios quiere)
		default:
			slog.Error(fmt.Sprintf("Algoritmo de planificacion no reconocido: %s", Config.ReadyIngressAlgorithm))
		}
	}
}

/*
func init() {
	go planificadorLargoPlazo()
	go planificadorCortoPlazo()
}
*/

// Mueve el pcb de una lista de procesos a otra EJ: mueve de NEW a READY y cambia al nuevo estado
func MoverPCB(pid uint, origen *[]structs.PCB, destino *[]structs.PCB, estadoNuevo string) {
	for i, pcb := range *origen {
		if pcb.PID == pid {
			pcb.Estado = estadoNuevo // cambiar el estado del PCB
			slog.Info(fmt.Sprintf("## (%d) pasa del estado %s al estado %s", pid, (*origen)[i].Estado, estadoNuevo))
			*destino = append(*destino, pcb)                    // mover a la cola destino
			*origen = append((*origen)[:i], (*origen)[i+1:]...) // eliminar del origen
			return
		}
	}
}

// ---------------------------- Funciones de prueba ----------------------------//
func NuevoProceso() structs.PCB {
	var pcb = CrearPCB(contadorProcesos)
	ColaNew = append(ColaNew, pcb)
	slog.Info(fmt.Sprintf("Se agregó el proceso %d a la cola de new", pcb.PID))
	contadorProcesos++
	return pcb
}

func CrearPCB(pid uint) structs.PCB {
	slog.Info(fmt.Sprintf("Se ha creado el proceso %d", pid))
	return structs.PCB{
		PID:            pid,
		PC:             0,
		Estado:         structs.EstadoNew,
		MetricasConteo: nil,
		MetricasTiempo: nil,
	}
}

//-------------------------------------------------------------------------------//

// ---------------------------- Funciones de test ----------------------------//
func TestCrearPCB() {
	// Crear múltiples PCBs para probar la variable global contadorProcesos
	pcb1 := NuevoProceso()
	if pcb1.PID != 0 {
		slog.Error("TestCrearPCB: El PID del primer proceso debería ser 0")
	}

	pcb2 := NuevoProceso()
	if pcb2.PID != 1 {
		slog.Error("TestCrearPCB: El PID del segundo proceso debería ser 1")
	}

	pcb3 := NuevoProceso()
	if pcb3.PID != 2 {
		slog.Error("TestCrearPCB: El PID del tercer proceso debería ser 2")
	}

	// Verificar otros atributos para cada PCB
	for _, pcb := range []structs.PCB{pcb1, pcb2, pcb3} {
		if pcb.PC != 0 {
			slog.Error(fmt.Sprintf("TestCrearPCB: El PC del proceso %d debería iniciar en 0", pcb.PID))
		}
		if pcb.Estado != structs.EstadoNew {
			slog.Error(fmt.Sprintf("TestCrearPCB: El estado inicial del proceso %d debería ser NEW", pcb.PID))
		}
	}

	slog.Info("TestCrearPCB completado")
}

func TestMoverAReady() {

	// Verificar que los procesos estén en NEW
	if len(ColaNew) != 3 {
		slog.Error("TestMoverAReady: Deberían haber 3 procesos en NEW")
	}

	// Mover procesos a READY
	MoverPCB(2, &ColaNew, &ColaReady, structs.EstadoReady)


	// Verificar que los procesos se movieron correctamente
	if len(ColaNew) != 2 {
		slog.Error("TestMoverAReady: La cola NEW debería haber 2 procesos")
	}

	if len(ColaReady) != 1 {
		slog.Error("TestMoverAReady: Deberían haber 1 procesos en READY")
	}

	// Verificar el estado de los procesos
	for _, pcb := range ColaReady {
		if pcb.Estado != structs.EstadoReady {
			slog.Error(fmt.Sprintf("TestMoverAReady: El proceso %d debería estar en estado READY", pcb.PID))
		}
	}

	slog.Info("TestMoverAReady completado")
}



func RunTests() {
	slog.Info("Iniciando tests...")
	TestCrearPCB()
	TestMoverAReady()
	slog.Info("Estado final de las colas:")
	slog.Info(fmt.Sprintf("Cola NEW: %+v", ColaNew))
	slog.Info(fmt.Sprintf("Cola READY: %+v", ColaReady))
	slog.Info("Tests completados")
}
