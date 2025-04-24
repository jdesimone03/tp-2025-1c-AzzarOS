package main

import (
	"cpu/utilsCPU"
	"utils"
)

func main() {
	utils.ConfigurarLogger("log_CPU")
	
	utils.EnviarMensaje(utilsCPU.Config.IPKernel, utilsCPU.Config.PortKernel,"peticionIO","Hola desde CPU")
	//utils.EnviarMensaje(config.IPMemory, config.PortMemory,"peticiones", "Hola desde CPU")
}