package main

import (
	"fmt"
	"net/http"
	"utils"
)

func main() {
	utils.ConfigurarLogger("log_KERNEL")
	config := utils.CargarConfiguracion[utils.ConfigKernel]("config.json")

	// estructura de prueba (reciclando la interfaz de io) para enviar mensajes a memoria

	interfaz := utils.Interfaz {
		Nombre: "socotroco",
		IP: config.IPMemory,
		Puerto: config.PortMemory,
	}


	utils.EnviarMensaje(config.IPMemory, config.PortMemory,"peticiones",interfaz)


	// mux.HandleFunc("/procesos", utils.RecibirPaquetes)
	http.HandleFunc("/interrupciones", utils.RecibirInterfaz)

	err := http.ListenAndServe(fmt.Sprintf(":%d",config.PortKernel), nil)
	if err != nil {
		panic(err)
	}
}
