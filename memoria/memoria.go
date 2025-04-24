package main

import (
	"memoria/utilsMemoria"
	"utils"
)

func main() {
	utils.ConfigurarLogger("log_MEMORIA")


	// http.HandleFunc("/peticiones", utils.RecibirInterfaz)
	utils.IniciarServidor(utilsMemoria.Config.PortMemory)
}
