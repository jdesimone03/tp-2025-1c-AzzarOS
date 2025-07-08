package main

import (
	"kernel/utilsKernel"
	"net/http"
	"os"
	"strconv"
	"utils"
	"utils/config"
	"utils/logueador"
)

func main() {

	// Carga los argumentos
	pscInicial := os.Args[1] //el pseudocodigo no va dentro de la memoria
	tamanioProceso, err := strconv.Atoi(os.Args[2])

	var rutaConfig string
    if len(os.Args) > 3 {
        rutaConfig = os.Args[3]
    } else {
        rutaConfig = "config/default.json"
    }

	// Inicia la configuración
	config.CargarConfiguracion(rutaConfig, &utilsKernel.Config)

	// Inicia el logueador
	logueador.ConfigurarLogger("log_KERNEL", utilsKernel.Config.LogLevel)
	
	if err != nil {
		logueador.Error("Error al convertir el tamaño del proceso a int")
		return
	}

	// Carga el proceso inicial
	utilsKernel.NuevoProceso(pscInicial, tamanioProceso) // memoria debe estar iniciada

	// Enter para iniciar el planificador de corto plazo
	utilsKernel.IniciarPlanificadores()

	// Handshakes
	http.HandleFunc("/handshake/CPU", utilsKernel.HandleHandshake("CPU"))
	http.HandleFunc("/handshake/IO", utilsKernel.HandleHandshake("IO"))

	// Syscalls
	http.HandleFunc("/syscall/IO", utilsKernel.HandleSyscall("IO"))
	http.HandleFunc("/syscall/INIT_PROC", utilsKernel.HandleSyscall("INIT_PROC"))
	http.HandleFunc("/syscall/DUMP_MEMORY", utilsKernel.HandleSyscall("DUMP_MEMORY"))
	http.HandleFunc("/syscall/EXIT", utilsKernel.HandleSyscall("EXIT"))

	//http.HandleFunc("/FinEjecucion",utilsKernel.Algo())
	http.HandleFunc("/guardar-contexto", utilsKernel.GuardarContexto)

	// Manejo de IO
	http.HandleFunc("/io-end", utilsKernel.HandleIOEnd)
	http.HandleFunc("/io-disconnect", utilsKernel.HandleIODisconnect)

	utils.IniciarServidor(utilsKernel.Config.PortKernel)
}
