package main

import (
	"memoria/utilsMemoria"
	"net/http"
	"os"
	"utils"
	"utils/config"
	"utils/logueador"
)

func main() {

	var rutaConfig string
    if len(os.Args) > 1 {
        rutaConfig = os.Args[1]
    } else {
        rutaConfig = "config/default.json"
    }

	// Inicia la configuración
	config.CargarConfiguracion(rutaConfig, &utilsMemoria.Config)
	
	// Inicia el logueador
	logueador.ConfigurarLogger("log_MEMORIA", utilsMemoria.Config.LogLevel)

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
	http.HandleFunc("/check-memoria", utilsMemoria.CheckMemoria) // para ver si hay espacio en memoria


	// GETS para mostrar información
	http.HandleFunc("/metricas", utilsMemoria.HandlerMostrarMetricas) // para ver las métricas de memoria
	http.HandleFunc("/listaProcEinstrucciones", utilsMemoria.HandlerMostrarProcesoConInstrucciones) // para ver los procesos y sus instrucciones
	http.HandleFunc("/ocupadas", utilsMemoria.MostrarOcupadas) // para ver los frames ocupados
	http.HandleFunc("/swap", utilsMemoria.HandlerMostrarSWAP) // para ver el contenido del swap
	http.HandleFunc("/memoriausuario", utilsMemoria.MostrarMemoria) // para ver los procesos en memoria

	
	// http.HandleFunc("/mover-a-swap", utilsMemoria.MoverProcesoASwap)

	utils.IniciarServidor(utilsMemoria.Config.PortMemory)
}
