package utilsKernel

import (
	"fmt"
	"net/http"
	"time"
	"utils/logueador"
	"utils/structs"
)

// Mueve el pcb de una lista de procesos a otra EJ: mueve de NEW a READY y cambia al nuevo estado
func MoverPCB(pid uint, origen *structs.ColaSegura, destino *structs.ColaSegura, estadoNuevo string) {
	pcb, indice := origen.Buscar(pid)
	if indice > -1 { // Si encuentra el indice
		estadoActual := pcb.Estado
		pcb.Estado = estadoNuevo // cambiar el estado del PCB

		pcb.MetricasTiempo[estadoActual] = time.Now().UnixMilli() - pcb.MetricasTiempo[estadoActual]
		if estadoNuevo != structs.EstadoExit {
			pcb.MetricasTiempo[estadoNuevo] = time.Now().UnixMilli()
		}

		pcb.MetricasConteo[estadoNuevo]++

		destino.Agregar(pcb)
		origen.Eliminar(indice)

		// Log obligatorio 3/8
		logueador.CambioDeEstado(pid, estadoActual, estadoNuevo)
		return
	}
}

func NuevoProceso(rutaArchInstrucciones string, tamanio int) {
	proceso := structs.NuevoProceso{
		PID:           contadorProcesos, // PID actual
		Instrucciones: rutaArchInstrucciones,
		Tamanio:       tamanio,
	}

	//utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "nuevo-proceso", proceso)
	NuevosProcesos.Agregar(proceso.PID, proceso)

	// Crea el PCB y lo inserta en NEW
	pcb := CrearPCB()
	pcb.MetricasConteo[structs.EstadoNew]++
	pcb.MetricasTiempo[structs.EstadoNew] = time.Now().UnixMilli()
	ColaNew.Agregar(pcb)
	contadorProcesos++
	TiempoEstimado.Agregar(pcb.PID,float64(Config.InitialEstimate))

	// Log obligatorio 2/8
	logueador.KernelCreacionDeProceso(pcb.PID)
}

func CrearPCB() structs.PCB {
	return structs.PCB{
		PID:            contadorProcesos,
		PC:             0,
		Estado:         structs.EstadoNew,
		MetricasConteo: make(map[string]int),
		MetricasTiempo: make(map[string]int64),
	}
}

func EstimarRafaga(pid uint) float64 {
	estimadoAnterior, _ := TiempoEstimado.Obtener(pid)
	tiempoEnExecute, _ := TiempoEnColaExecute.Obtener(pid)
	realAnterior := time.Now().UnixMilli() - tiempoEnExecute
	return float64(realAnterior)*Config.Alpha + (1-Config.Alpha)*estimadoAnterior
}

func Interrumpir(nombreCpu string) {
	cpu, existe := InstanciasCPU.Obtener(nombreCpu)
	if !existe {
		logueador.Error("No se pudo interrumpir %s ya que no existe en el sistema.", nombreCpu)
		return
	}

	url := fmt.Sprintf("http://%s:%d/interrupt", cpu.IP, cpu.Puerto)
	_, err := http.Get(url)

	if err != nil {
		logueador.Error("No se pudo interrumpir la CPU %s: %v", nombreCpu, err)
	}
	<-chCambioDeContexto
	logueador.Info("supuestamente ya se hizo el cambio de contexto")
}

func IniciarTimerSuspension(pid uint) {
	// Crear nuevo timer
	if _, exists := TiempoEnColaBlocked[pid]; exists {
		logueador.Warn("Intento de iniciar un timer para PID %d que ya tiene uno activo. Ignorando.", pid)
		return
	}
	timer := time.NewTimer(time.Duration(Config.SuspensionTime) * time.Millisecond)
	TiempoEnColaBlocked[pid] = timer

	// Log de inicio del timer
	logueador.Info("Timer de suspensión configurado para PID %d. Duración: %d ms. Será evaluado por PlanificadorMedianoPlazo.", pid, Config.SuspensionTime)
	// La lógica de expiración y movimiento ahora es manejada por PlanificadorMedianoPlazo.
}
