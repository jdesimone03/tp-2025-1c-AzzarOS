package main

import (
	"cpu/utilsCPU"
	"fmt"
	"net/http"
	"os"
	"utils"
	"utils/structs"
)

func main() {
	identificador := os.Args[1]

	utils.ConfigurarLogger(fmt.Sprintf("log_CPU_%s",identificador))

	cpu := structs.CPU{
		IP:     utilsCPU.Config.IPCPU,
		Puerto: utilsCPU.Config.PortCPU,
	}
	
	peticion := structs.HandshakeCPU{
		Identificador: identificador,
		CPU: cpu,
	}
	
	utils.EnviarMensaje(utilsCPU.Config.IPKernel, utilsCPU.Config.PortKernel,"handshake/CPU",peticion)

	http.HandleFunc("/ejecutar", utilsCPU.RecibirEjecucion)
	//http.HandleFunc("/interrupciones", utilsCPU.RecibirPeticion)

}

