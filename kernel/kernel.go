package main

import (
	"kernel/utilsKernel"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"utils"
)

func main() {
	pscInicial := os.Args[1] //el pseudocodigo no va dentro de la memoria

	tamanioProceso, err := strconv.Atoi(os.Args[2])
	if err != nil {
		slog.Error("Error al convertir el tamaño del proceso a int")
		return
	}

	utils.ConfigurarLogger("log_KERNEL")

	// memoria debe estar iniciada
	utilsKernel.NuevoProceso(pscInicial, tamanioProceso)
	
	// Enter para iniciar el planificador
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


	utils.IniciarServidor(utilsKernel.Config.PortKernel)

}
