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
	utilsIO.NombreInterfaz = nombre

	// Inicia el logueador
	logueador.ConfigurarLogger(fmt.Sprintf("log_IO_%s", nombre))

	// Inicia la configuración
	config.CargarConfiguracion("config.json", &utilsIO.Config)

	interfaz := structs.InterfazIO{
		IP:     utilsIO.Config.IPIo,
		Puerto: utilsIO.Config.PortIo,
	}

	peticion := structs.HandshakeIO{
		Nombre:   nombre,
		Interfaz: interfaz,
	}

	utils.EnviarMensaje(utilsIO.Config.IPKernel, utilsIO.Config.PortKernel, "handshake/IO", peticion)

	http.HandleFunc("/ejecutarIO", utilsIO.RecibirEjecucionIO)

	utils.IniciarServidor(utilsIO.Config.PortIo)
}
