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

	mux.HandleFunc("/inicializarProceso", utilsMemoria.HandlerDePedidoDeInicializacion) // CPU -> KERNEL -> MEMORIA
	mux.HandleFunc("/proximaInstruccion", utilsMemoria.HandlerPedidoDeInstruccion) // CPU -> MEMORIA
	mux.HandleFunc("/memoryDump", utilsMemoria.HandlerMEMORYDUMP) // KERNEL -> MEMORIA
	mux.HandleFunc("/write", utilsMemoria.HandlerWrite) // CPU -> MEMORIA
	mux.HandleFunc("/read", utilsMemoria.HandlerRead) // CPU -> MEMORIA
	mux.HandleFunc("/suspenderProceso", utilsMemoria.HandlerDeSuspension) // CPU -> MEMORIA
	mux.HandleFunc("/desuspenderProceso", utilsMemoria.HandlerDeDesuspension) // CPU -> MEMORIA
	mux.HandleFunc("/finalizarProceso", utilsMemoria.HandlerDeFinalizacion) // CPU -> MEMORIA
	mux.HandleFunc("/actualizarMP", utilsMemoria.HandlerEscribirDeCache) // CPU actualiza la MP con la caché 
	mux.HandleFunc("/config", utilsMemoria.HandlerConfig) // para que CPU sepa de la configuración de memoria
	mux.HandleFunc("/tabla-paginas", utilsMemoria.HandlerPedidoTDP) // para caché-CPU
	mux.HandleFunc("/pedirFrame", utilsMemoria.HandlerPedidoFrame) // para caché-CPU


	// GETS para mostrar información
	mux.HandleFunc("/metricas", utilsMemoria.HandlerMostrarMetricas)
	mux.HandleFunc("/listaProcEinstrucciones", utilsMemoria.HandlerMostrarProcesoConInstrucciones)
	mux.HandleFunc("/ocupadas", utilsMemoria.MostrarOcupadas)
	
	// http.HandleFunc("/mover-a-swap", utilsMemoria.MoverProcesoASwap)

	utils.IniciarServidor(utilsMemoria.Config.PortMemory)
}
