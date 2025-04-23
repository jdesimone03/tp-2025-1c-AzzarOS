package main

import (
	"os"
	"utils"
)

// TODO levantar servidor de IO despues del handshake
func main() {
	utils.ConfigurarLogger("log_IO")

	config := utils.CargarConfiguracion[utils.ConfigIO]("config.json")

	nombre := os.Args[1]

	interfaz := utils.Interfaz {
		Nombre: nombre,
		IP: config.IPIo,
		Puerto: config.PortIo,
	}

	utils.EnviarMensaje(config.IPKernel, config.PortKernel, "interrupciones", interfaz)


}
