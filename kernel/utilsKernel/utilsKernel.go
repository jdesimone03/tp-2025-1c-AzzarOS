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
var ColaExecute []structs.PCB
var ColaBlocked []structs.PCB
var ColaExit []structs.PCB

// scheduler_algorithm: LARGO plazo
// ready_ingress_algorithm: CORTO plazo

var InstanciasCPU = make(map[string]structs.CPU)

var ListaExecIO = make(map[string]uint) // nombre de io: PID
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

			Interfaces[interfaz.Nombre] = interfaz.Interfaz

			aux, encontrado := ListaWaitIO[interfaz.Nombre]
			if encontrado {
				// Borra el primer elemento en la lista de espera
				aEjecutar := aux[0]
				aux := append(aux[:0], aux[1:]...)
				ListaWaitIO[interfaz.Nombre] = aux

				peticion := structs.PeticionIO{
					PID:            aEjecutar.PID,
					NombreIfaz:     interfaz.Nombre,
					SuspensionTime: aEjecutar.TiempoMs,
				}
				utils.EnviarMensaje(interfaz.Interfaz.IP, interfaz.Interfaz.Puerto, "/peticionIO", peticion)
				ListaExecIO[interfaz.Nombre] = aEjecutar.PID
			}
			slog.Info(fmt.Sprintf("Me llego la siguiente interfaz: %+v", interfaz))

		case "CPU":
			instancia, err := utils.DecodificarMensaje[structs.PeticionCPU](r)
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
			peticion, err := utils.DecodificarMensaje[structs.PeticionIO](r)
			if err != nil {
				slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Error al decodificar mensaje"))
				return
			}

			pid := peticion.PID
			nombre := peticion.NombreIfaz
			tiempoMs := peticion.SuspensionTime

			interfaz, encontrada := Interfaces[nombre]
			if encontrada {
				_, isExec := ListaExecIO[nombre]
				if isExec {
					// Enviar proceso a BLOCKED
					MoverPCB(pid, &ColaExecute, &ColaBlocked, structs.EstadoBlocked)

					// Enviar proceso a ListaWaitIO
					espera := structs.EsperaIO{
						PID:      pid,
						TiempoMs: tiempoMs,
					}
					ListaWaitIO[nombre] = append(ListaWaitIO[nombre], espera)
				} else {
					// Enviar al IO el PID y el tiempo en ms
					peticion := structs.PeticionIO{
						PID:            pid,
						NombreIfaz:     nombre,
						SuspensionTime: tiempoMs,
					}
					ListaExecIO[nombre] = pid
					utils.EnviarMensaje(interfaz.IP, interfaz.Puerto, "peticionIO", peticion)
				}
			} else {
				slog.Error(fmt.Sprintf("La interfaz %s no existe en el sistema", nombre))

				// Enviar proceso a EXIT
				MoverPCB(pid, &ColaExecute, &ColaExit, structs.EstadoBlocked)
			}
		default:
			slog.Error(fmt.Sprintf("FATAL: %s no es un tipo de syscall valida.", tipo))
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)

	}
	// Despues le hacemos un case switch para cada syscall diferente
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
func NuevoProceso(pid uint) structs.PCB {
	var pcb = CrearPCB(pid, 0, structs.EstadoNew)
	ColaNew = append(ColaNew, pcb)
	slog.Info(fmt.Sprintf("Se agregó el proceso %d a la cola de new", pcb.PID))
	return pcb
}

func CrearPCB(pid uint, pc uint, estado string) structs.PCB {
	slog.Info(fmt.Sprintf("Se ha creado el proceso %d", pid))
	return structs.PCB{
		PID:            pid,
		PC:             pc,
		Estado:         estado,
		MetricasConteo: nil,
		MetricasTiempo: nil,
	}
}

//-------------------------------------------------------------------------------//
