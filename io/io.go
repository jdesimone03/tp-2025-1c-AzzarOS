package main

import (
	"utils"
)

func main() {
	utils.ConfigurarLogger("log_IO")
	config := utils.CargarConfiguracion[utils.ConfigIO]("config.json")
	utils.EnviarMensaje(config.IPKernel, config.PortKernel, "interrupciones", "Hola desde IO")
}
