package main

import (
	"cpu/utilsCPU"
	"net/http"
	"os"
	"utils"
	"utils/config"
	"utils/logueador"
	"utils/structs"
)

func main() {
	// Carga los argumentos
	identificador := os.Args[1]

	var rutaConfig string
    if len(os.Args) > 2 {
        rutaConfig = os.Args[2]
    } else {
        rutaConfig = "config/default.json"
    }
	
	// Inicia la configuración
	config.CargarConfiguracion(rutaConfig, &utilsCPU.Config)

	// Inicia el logueador
	logueador.ConfigurarLogger("log_CPU_" + identificador, utilsCPU.Config.LogLevel)
	
	// Cargar la configuración de memoria
	err := utilsCPU.PedirConfigMemoria()
	if err != nil {
		logueador.Error("No se pudo obtener la configuración de memoria: %v", err)
		return
	}

	cpu := structs.InstanciaCPU{
		Nombre:		identificador,
		IP:         utilsCPU.Config.IPCPU,
		Puerto:     utilsCPU.Config.PortCPU,
	}

	peticion := structs.HandshakeCPU{
		Identificador: identificador,
		CPU:           cpu,
	}

	utils.EnviarMensaje(utilsCPU.Config.IPKernel, utilsCPU.Config.PortKernel, "handshake/CPU", peticion)

	utilsCPU.InicializarCache()
	utilsCPU.InicializarTLB()

	utilsCPU.MostrarContenidoTLB()

	http.HandleFunc("/dispatch", utilsCPU.RecibirEjecucion)
	http.HandleFunc("/interrupt", utilsCPU.RecibirInterrupcion)
	http.HandleFunc("/tlb", utilsCPU.MostrarTLB)
	http.HandleFunc("/cache", utilsCPU.MostrarCache)
	http.HandleFunc("/desalojo", utilsCPU.DesaolojoDeProceso) // El en query tiene que venir el PID del proceso a desalojar

	utils.IniciarServidor(utilsCPU.Config.PortCPU)
}
