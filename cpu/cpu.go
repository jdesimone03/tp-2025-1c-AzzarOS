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
	
	peticion := structs.PeticionCPU{
		Identificador: identificador,
		CPU: cpu,
	}
	
	utils.EnviarMensaje(utilsCPU.Config.IPKernel, utilsCPU.Config.PortKernel,"handshakeCPU",peticion)

	http.HandleFunc("/ejecutar", utilsCPU.RecibirEjecucion)
	//http.HandleFunc("/interrupciones", utilsCPU.RecibirPeticion)
	//http.HandleFunc("/fetch", utilsCPU.DecodificarInstruccion)

	//utils.EnviarMensaje(config.IPMemory, config.PortMemory,"peticiones", "Hola desde CPU")
}