package main

import (
	"fmt"
	"log"
	"net/http"
	"utils"
)

func main() {
	utils.ConfigurarLogger("log_KERNEL")
	kernelConfig := utils.CargarConfiguracion[utils.ConfigKernel]("config.json")

	if kernelConfig == nil {
		log.Println("Error al cargar la configuracion de Kernel")
		return
	}

	utils.EnviarMensaje(kernelConfig.IPMemory, kernelConfig.PortMemory,"peticiones","Hola desde Kernel")

	mux := http.NewServeMux()

	// mux.HandleFunc("/procesos", utils.RecibirPaquetes)
	mux.HandleFunc("/interrupciones", utils.RecibirMensaje)

	err := http.ListenAndServe(fmt.Sprintf(":%d",kernelConfig.PortKernel), mux)
	if err != nil {
		panic(err)
	}
}
