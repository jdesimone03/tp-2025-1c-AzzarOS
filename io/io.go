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

	var rutaConfig string
	if len(os.Args) > 2 {
		rutaConfig = os.Args[2]
	} else {
		rutaConfig = "config/default.json"
	}

	// Inicia la configuración
	config.CargarConfiguracion(rutaConfig, &utilsIO.Config)

	// Inicia el logueador
	logueador.ConfigurarLogger("log_IO_"+nombre, utilsIO.Config.LogLevel)

	http.HandleFunc("/ejecutarIO", utilsIO.RecibirEjecucionIO)

	// Canal para confirmar que el servidor está listo
	servidorListo := make(chan bool)

	// Inicia el servidor en una goroutine
	go func() {
		servidorListo <- true // Señala que está listo para iniciar
		utils.IniciarServidor(utilsIO.Config.PortIo)
	}()

	// Espera confirmación de que el servidor está iniciando
	<-servidorListo

	interfaz := structs.InterfazIO{
		Nombre:     nombre,
		IP:         utilsIO.Config.IPIo,
		Puerto:     utilsIO.Config.PortIo,
	}

	peticion := structs.HandshakeIO{
		Nombre:   nombre,
		Interfaz: interfaz,
	}

	utilsIO.Interfaz = interfaz

	utils.EnviarMensaje(utilsIO.Config.IPKernel, utilsIO.Config.PortKernel, "handshake/IO", peticion)

	//utils.IniciarServidor(utilsIO.Config.PortIo)

	select {}
}
