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

	// Cargo las estructuras
	utilsMemoria.IniciarEstructuras()

	// Endpoints
	http.HandleFunc("/fetch", utilsMemoria.EnviarInstruccion)
	http.HandleFunc("/nuevo-proceso", utilsMemoria.NuevoProceso)
	http.HandleFunc("/check-memoria", utilsMemoria.CheckMemoria)

	http.HandleFunc("/read", utilsMemoria.Read)
	http.HandleFunc("/write", utilsMemoria.Write)

	http.HandleFunc("/mover-a-swap", utilsMemoria.MoverProcesoASwap)

	utils.IniciarServidor(utilsMemoria.Config.PortMemory)
}
