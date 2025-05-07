package main

import (
	"kernel/utilsKernel"
	"net/http"
	"os"
	"strconv"
	"utils"
	"log/slog"
)

func main() {
	pscInicial := os.Args[1] //el pseudocodigo no va dentro de la memoria

	tamanioProceso, err := strconv.Atoi(os.Args[2])
	if err != nil {
		slog.Error("Error al convertir el tamaño del proceso a int")
		return
	}

	utils.ConfigurarLogger("log_KERNEL")

	utilsKernel.NuevoProceso(pscInicial, tamanioProceso)//esto se manda a memoria


	// Handshakes
	http.HandleFunc("/handshakeCPU", utilsKernel.HandleHandshake("CPU"))
	http.HandleFunc("/handshakeIO", utilsKernel.HandleHandshake("IO"))

	http.HandleFunc("/syscallIO", utilsKernel.HandleSyscall("IO"))
	http.HandleFunc("/syscallInitProc", utilsKernel.HandleSyscall("INIT_PROC"))

	utils.IniciarServidor(utilsKernel.Config.PortKernel)
}
