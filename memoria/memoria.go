package main

import (
	"memoria/utilsMemoria"
	"net/http"
	"utils"
)

func main() {
	utils.ConfigurarLogger("log_MEMORIA")

	utils.IniciarServidor(utilsMemoria.Config.PortMemory)

	http.HandleFunc("/fetch", utilsMemoria.EnviarInstruccion)
	http.HandleFunc("/nuevo-proceso", utilsMemoria.NuevoProceso)
	http.HandleFunc("/check-memoria", utilsMemoria.CheckMemoria)
}
