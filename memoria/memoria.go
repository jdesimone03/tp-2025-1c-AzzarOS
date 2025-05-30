package main

import (
	"memoria/utilsMemoria"
	"net/http"
	"utils"
	"utils/config"
	"utils/logueador"
)

func main() {
	logueador.ConfigurarLogger("log_MEMORIA")
	
	// Inicia la configuración
	config.CargarConfiguracion("config.json", &utilsMemoria.Config)

	http.HandleFunc("/fetch", utilsMemoria.EnviarInstruccion)
	http.HandleFunc("/nuevo-proceso", utilsMemoria.NuevoProceso)
	http.HandleFunc("/check-memoria", utilsMemoria.CheckMemoria)

	utils.IniciarServidor(utilsMemoria.Config.PortMemory)

}
