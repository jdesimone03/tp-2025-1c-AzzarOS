package main

import (
	"utils"
)

func main() {
	utils.ConfigurarLogger("log_CPU")
	config := utils.CargarConfiguracion[utils.ConfigCPU]("config.json")

	utils.EnviarMensaje(config.IPKernel, config.PortKernel,"interrupciones","Hola desde CPU")
	utils.EnviarMensaje(config.IPMemory, config.PortMemory,"peticiones", "Hola desde CPU")
}