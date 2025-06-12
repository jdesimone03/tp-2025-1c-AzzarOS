package utilsKernel

import (
	"fmt"
	"net/http"
	"time"
	"utils"
	"utils/logueador"
	"utils/structs"
)

// Mueve el pcb de una lista de procesos a otra EJ: mueve de NEW a READY y cambia al nuevo estado
func MoverPCB(pid uint, origen *structs.ColaSegura, destino *structs.ColaSegura, estadoNuevo string) {
	pcb, indice := origen.Buscar(pid)
	if indice > -1 { // Si encuentra el indice
		estadoActual := pcb.Estado
		pcb.Estado = estadoNuevo // cambiar el estado del PCB

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
	NuevosProcesos[proceso.PID] = proceso

	// Crea el PCB y lo inserta en NEW
	pcb := CrearPCB()
	pcb.MetricasConteo[structs.EstadoNew]++
	ColaNew.Agregar(pcb)
	contadorProcesos++

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

func GetCPUDisponible() (string, bool) {
	for nombre, valores := range InstanciasCPU {
		if !valores.Ejecutando {
			return nombre, true
		}
	}
	return "", false
}

func GetCPU(pid uint) string {
	for nombre, valores := range InstanciasCPU {
		if valores.PID == pid {
			return nombre
		}
	}
	return ""
}

func EstimarRafaga(estimadoAnterior float64, realAnterior float64) float64 {
	return realAnterior*Config.Alpha + (1-Config.Alpha)*estimadoAnterior
}

func RecibirTiempoEjecucion(w http.ResponseWriter, r *http.Request) {
	tiempo, err := utils.DecodificarMensaje[structs.TiempoEjecucion](r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	TiempoEnColaExecute[tiempo.PID] = tiempo.Tiempo
	w.WriteHeader(http.StatusOK)
}

func Interrumpir(nombreCpu string) {
	cpu, existe := InstanciasCPU[nombreCpu]
	if !existe {
		logueador.Error("No se pudo interrumpir %s ya que no existe en el sistema.", nombreCpu)
		return
	}

	url := fmt.Sprintf("http://%s:%d/interrupt", cpu.IP, cpu.Puerto)
	_, err := http.Get(url)

	if err != nil {
		logueador.Error("No se pudo interrumpir la CPU %s: %v", nombreCpu, err)
	}

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