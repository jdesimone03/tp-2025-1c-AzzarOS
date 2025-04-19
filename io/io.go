package main

import (
	"fmt"
	"os"
	"utils"
)


func main() {
	nombre := os.Args[1]
	fmt.Println(nombre)
	utils.ConfigurarLogger("log_IO")
	config := utils.CargarConfiguracion[utils.ConfigIO]("config.json")
	interfaz := utils.Interfaz {
		Nombre: nombre,
		IP: config.IPIo,
		Puerto: config.PortIo,
	}

	utils.EnviarMensaje(config.IPKernel, config.PortKernel, "interrupciones", interfaz)
}
