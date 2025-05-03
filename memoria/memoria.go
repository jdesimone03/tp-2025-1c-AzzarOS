package main

import (
	"memoria/utilsMemoria"
	"net/http"
	"utils"
)

func main() {
	utils.ConfigurarLogger("log_MEMORIA")


	// http.HandleFunc("/peticiones", utils.RecibirInterfaz)
	utils.IniciarServidor(utilsMemoria.Config.PortMemory)

	http.HandleFunc("/fetch", utilsMemoria.EnviarInstruccion)
}
