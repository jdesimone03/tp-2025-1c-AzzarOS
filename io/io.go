package main

import (
	"utils"
	"log"
)

func main() {
	utils.ConfigurarLogger("log_IO")
	IOConfig := utils.CargarConfiguracion[utils.ConfigIO]("config.json")

	if IOConfig == nil {
		log.Println("Error al cargar la configuracion de IO")
		return
	}

	log.Printf("Configuracion de IO cargada correctamente: %+v", IOConfig)

	utils.EnviarMensaje(IOConfig.IPKernel, IOConfig.PortKernel,"interrupciones", "Hola desde IO")

	
}
