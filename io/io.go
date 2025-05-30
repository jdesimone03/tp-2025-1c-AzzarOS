package main

import (
	"net/http"
	"os"
	"tp/io/utilsIO"
	"utils"
	"utils/structs"
)

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

	utils.EnviarMensaje(utilsIO.Config.IPKernel, utilsIO.Config.PortKernel, "handshake/IO", peticion)

	http.HandleFunc("/ejecutarIO", utilsIO.RecibirPeticion)
	http.HandleFunc("/ping", utilsIO.Ping)

	utils.IniciarServidor(utilsIO.Config.PortIo)
}
