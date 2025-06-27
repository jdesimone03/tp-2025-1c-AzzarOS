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
	http.HandleFunc("/inicializarProceso", utilsMemoria.HandlerDePedidoDeInicializacion) // CPU -> KERNEL -> MEMORIA
	http.HandleFunc("/proximaInstruccion", utilsMemoria.HandlerPedidoDeInstruccion) // CPU -> MEMORIA
	http.HandleFunc("/memoryDump", utilsMemoria.HandlerMEMORYDUMP) // KERNEL -> MEMORIA
	http.HandleFunc("/write", utilsMemoria.HandlerWrite) // CPU -> MEMORIA
	http.HandleFunc("/read", utilsMemoria.HandlerRead) // CPU -> MEMORIA
	http.HandleFunc("/suspenderProceso", utilsMemoria.HandlerDeSuspension) // CPU -> MEMORIA
	http.HandleFunc("/desuspenderProceso", utilsMemoria.HandlerDeDesuspension) // CPU -> MEMORIA
	http.HandleFunc("/finalizarProceso", utilsMemoria.HandlerDeFinalizacion) // CPU -> MEMORIA
	http.HandleFunc("/actualizarMP", utilsMemoria.HandlerEscribirDeCache) // CPU actualiza la MP con la caché 
	http.HandleFunc("/config", utilsMemoria.HandlerConfig) // para que CPU sepa de la configuración de memoria
	http.HandleFunc("/tabla-paginas", utilsMemoria.HandlerPedidoTDP) // para caché-CPU
	http.HandleFunc("/pedirFrame", utilsMemoria.HandlerPedidoFrame) // para caché-CPU


	// GETS para mostrar información
	http.HandleFunc("/metricas", utilsMemoria.HandlerMostrarMetricas)
	http.HandleFunc("/listaProcEinstrucciones", utilsMemoria.HandlerMostrarProcesoConInstrucciones)
	http.HandleFunc("/ocupadas", utilsMemoria.MostrarOcupadas)
	http.HandleFunc("/swap", utilsMemoria.HandlerMostrarSWAP) // para ver el contenido del swap
	

	
	// http.HandleFunc("/mover-a-swap", utilsMemoria.MoverProcesoASwap)

	utils.IniciarServidor(utilsMemoria.Config.PortMemory)
}
