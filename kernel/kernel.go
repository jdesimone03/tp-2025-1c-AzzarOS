package main

import (
	"fmt"
	"net/http"
	"utils"
)

func main() {
	utils.ConfigurarLogger("log_KERNEL")
	config := utils.CargarConfiguracion[utils.ConfigKernel]("config.json")

	utils.EnviarMensaje(config.IPMemory, config.PortMemory,"peticiones","Hola desde Kernel")

	mux := http.NewServeMux()

	// mux.HandleFunc("/procesos", utils.RecibirPaquetes)
	mux.HandleFunc("/interrupciones", utils.RecibirInterfaz)

	err := http.ListenAndServe(fmt.Sprintf(":%d",config.PortKernel), mux)
	if err != nil {
		panic(err)
	}
}
