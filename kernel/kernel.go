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
	http.HandleFunc("/handshake/CPU", utilsKernel.HandleHandshake("CPU"))
	http.HandleFunc("/handshake/IO", utilsKernel.HandleHandshake("IO"))

	// Syscalls
	http.HandleFunc("/syscall/IO", utilsKernel.HandleSyscall("IO"))
	http.HandleFunc("/syscall/INIT_PROC", utilsKernel.HandleSyscall("INIT_PROC"))
	http.HandleFunc("/syscall/DUMP_MEMORY", utilsKernel.HandleSyscall("DUMP_MEMORY"))
	http.HandleFunc("/syscall/EXIT", utilsKernel.HandleSyscall("EXIT"))

	utils.IniciarServidor(utilsKernel.Config.PortKernel)
}
