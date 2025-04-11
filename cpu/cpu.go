package main

import (
	"utils"
	"log"
)

func main() {
	utils.ConfigurarLogger("log_CPU")
	CPUConfig := utils.CargarConfiguracion[utils.ConfigCPU]("config.json")

	if CPUConfig == nil {
		log.Println("Error al cargar la configuracion de CPU")
		return
	}

	log.Printf("Configuracion de CPU cargada correctamente: %+v", CPUConfig)

	utils.EnviarMensaje(CPUConfig.IPKernel, CPUConfig.PortKernel, "Hola desde CPU")


}
