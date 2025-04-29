package main

import (
	"kernel/utilsKernel"
	"net/http"
	"utils"
)

// TODO crear variable global que guarde la interfaz actual
func main() {
	//pseudocodigo := os.Args[1] // el pseudocodigo no va dentro de la memoria
	//tam_proceso := os.Args[2]
	//esto se manda a memoria
	utils.ConfigurarLogger("log_KERNEL")

	// Handshakes
	http.HandleFunc("/handshakeCPU", utilsKernel.HandleHandshake("CPU"))
	http.HandleFunc("/handshakeIO", utilsKernel.HandleHandshake("IO"))

	http.HandleFunc("/syscallIO", utilsKernel.HandleSyscall("IO")) // podria ser handlerSyscall

	utils.IniciarServidor(utilsKernel.Config.PortKernel)
}
