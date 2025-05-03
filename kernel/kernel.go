package main

import (
	"kernel/utilsKernel"
	"net/http"
	// "os"
	"utils"
)

func main() {
	// pscInicial := os.Args[1] //el pseudocodigo no va dentro de la memoria
	// tamanioProceso := os.Args[2]
	//esto se manda a memoria
	utils.ConfigurarLogger("log_KERNEL")

	// Handshakes
	http.HandleFunc("/handshakeCPU", utilsKernel.HandleHandshake("CPU"))
	http.HandleFunc("/handshakeIO", utilsKernel.HandleHandshake("IO"))

	http.HandleFunc("/syscallIO", utilsKernel.HandleSyscall("IO"))
	http.HandleFunc("/syscallInitProc", utilsKernel.HandleSyscall("INIT_PROC"))

	utils.IniciarServidor(utilsKernel.Config.PortKernel)
}
