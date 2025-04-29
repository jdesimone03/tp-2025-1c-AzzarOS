package main

import (
	"net/http"
	"os"
	"tp/io/utilsIO"
	"utils"
	"utils/structs"
)

// TODO levantar servidor de IO despues del handshake
func main() {
	utils.ConfigurarLogger("log_IO")

	nombre := os.Args[1]

	interfaz := structs.Interfaz{
		IP:     utilsIO.Config.IPIo,
		Puerto: utilsIO.Config.PortIo,
	}

	peticion := structs.HandshakeIO{
		Nombre:   nombre,
		Interfaz: interfaz,
	}

	utils.EnviarMensaje(utilsIO.Config.IPKernel, utilsIO.Config.PortKernel, "handshakeIO", peticion)

	http.HandleFunc("/peticionIO", utilsIO.RecibirPeticion)

	utils.IniciarServidor(utilsIO.Config.PortIo)
}
