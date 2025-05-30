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

	cpu := structs.CPU{
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
	http.HandleFunc("/ping", utilsCPU.PingCPU)

	utils.IniciarServidor(utilsCPU.Config.PortCPU)
}
