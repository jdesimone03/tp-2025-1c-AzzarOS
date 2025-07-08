package main

import (
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

	var rutaConfig string
	if len(os.Args) > 2 {
        rutaConfig = os.Args[2]
    } else {
        rutaConfig = "config/default.json"
    }

	// Inicia la configuración
	config.CargarConfiguracion(rutaConfig, &utilsIO.Config)
	
	// Inicia el logueador
	logueador.ConfigurarLogger("log_IO_" + nombre, utilsIO.Config.LogLevel)

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
