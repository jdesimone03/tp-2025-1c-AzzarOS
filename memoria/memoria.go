package main

import (
	"memoria/utilsMemoria"
	"net/http"
	"utils"
	"utils/logueador"
)

func main() {
	logueador.ConfigurarLogger("log_MEMORIA")

	http.HandleFunc("/fetch", utilsMemoria.EnviarInstruccion)
	http.HandleFunc("/nuevo-proceso", utilsMemoria.NuevoProceso)
	http.HandleFunc("/check-memoria", utilsMemoria.CheckMemoria)

	utils.IniciarServidor(utilsMemoria.Config.PortMemory)

}
