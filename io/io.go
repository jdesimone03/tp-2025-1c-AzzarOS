package main

import (
	"fmt"
	"net/http"
	"os"
	"tp/io/utilsIO"
	"utils"
	"utils/config"
	"utils/logueador"
	"utils/structs"
)

func main() {
	// Carga los argumentos
	nombre := os.Args[1]

	// Inicia el logueador
	logueador.ConfigurarLogger(fmt.Sprintf("log_IO_%s", nombre))

	// Inicia la configuración
	config.CargarConfiguracion("config.json", &utilsIO.Config)

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
