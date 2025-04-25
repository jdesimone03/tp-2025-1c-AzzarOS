package main

import (
	"cpu/utilsCPU"
	"fmt"
	"os"
	"utils"
	"utils/structs"
)

func main() {
	identificador := os.Args[1]

	utils.ConfigurarLogger(fmt.Sprintf("log_CPU_%s",identificador))

	cpu := structs.CPU{
		IP:     utilsCPU.Config.IPCPU,
		Puerto: utilsCPU.Config.PortCPU,
	}
	
	peticion := structs.PeticionCPU{
		Identificador: identificador,
		CPU: cpu,
	}
	
	utils.EnviarMensaje(utilsCPU.Config.IPKernel, utilsCPU.Config.PortKernel,"handshakeCPU",peticion)
	//utils.EnviarMensaje(config.IPMemory, config.PortMemory,"peticiones", "Hola desde CPU")
}