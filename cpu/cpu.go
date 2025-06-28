package main

import (
	"cpu/utilsCPU"
	"fmt"
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

	// Inicia el logueador
	logueador.ConfigurarLogger(fmt.Sprintf("log_CPU_%s", identificador))

	// Inicia la configuración
	config.CargarConfiguracion("config.json", &utilsCPU.Config)

	// Cargar la configuración de memoria
	err := utilsCPU.PedirConfigMemoria()
	if err != nil {
		logueador.Error("No se pudo obtener la configuración de memoria: %v", err)
		return
	}

	cpu := structs.InstanciaCPU{
		IP:         utilsCPU.Config.IPCPU,
		Puerto:     utilsCPU.Config.PortCPU,
		Ejecutando: false,
	}

	peticion := structs.HandshakeCPU{
		Identificador: identificador,
		CPU:           cpu,
	}

	utils.EnviarMensaje(utilsCPU.Config.IPKernel, utilsCPU.Config.PortKernel, "handshake/CPU", peticion)

	http.HandleFunc("/dispatch", utilsCPU.RecibirEjecucion)
	http.HandleFunc("/interrupt", utilsCPU.RecibirInterrupcion)
	http.HandleFunc("/tlb", utilsCPU.MostrarTLB)
	http.HandleFunc("/cache", utilsCPU.MostrarCache)

	utils.IniciarServidor(utilsCPU.Config.PortCPU)
}
