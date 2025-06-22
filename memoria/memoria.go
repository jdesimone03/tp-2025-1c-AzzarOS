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

	// Mux 
	mux := http.NewServeMux()

	// Endpoints

	mux.HandleFunc("/inicializarProceso", utilsMemoria.HandlerDePedidoDeInicializacion)
	mux.HandleFunc("/proximaInstruccion", utilsMemoria.HandlerPedidoDeInstruccion)
	mux.HandleFunc("/memoryDump", utilsMemoria.HandlerMEMORYDUMP)
	mux.HandleFunc("/write", utilsMemoria.HandlerWrite)
	mux.HandleFunc("/read", utilsMemoria.HandlerRead)
	mux.HandleFunc("/suspenderProceso", utilsMemoria.HandlerDeSuspension)
	mux.HandleFunc("/desuspenderProceso", utilsMemoria.HandlerDeDesuspension)
	mux.HandleFunc("/finalizarProceso", utilsMemoria.HandlerDeFinalizacion)

	// GETS para mostrar información
	mux.HandleFunc("/metricas", utilsMemoria.HandlerMostrarMetricas)
	mux.HandleFunc("/listaProcEinstrucciones", utilsMemoria.HandlerMostrarProcesoConInstrucciones)
	mux.HandleFunc("/tablasDePaginas", utilsMemoria.MostrarTablasDePaginas)
	mux.HandleFunc("/ocupadas", utilsMemoria.MostrarOcupadas)
	
	// http.HandleFunc("/mover-a-swap", utilsMemoria.MoverProcesoASwap)

	utils.IniciarServidor(utilsMemoria.Config.PortMemory)
}
