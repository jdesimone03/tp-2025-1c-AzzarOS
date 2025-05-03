package utilsCPU

import (
	"fmt"
	"log/slog"
	"net/http"
	"utils"
	"utils/structs"
	"utils/config"
)

var Config = config.CargarConfiguracion[config.ConfigCPU]("config.json")

func RecibirEjecucion(w http.ResponseWriter, r *http.Request) {
	_, err := utils.DecodificarMensaje[structs.PCB](r) //despues la variable le pongo pcb para que se pueda manipular, sino llora el lenguaje
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	// Solicitar a memoria la siguiente instruccion para la ejecución
	// InstruccionCodificada = FETCH(PCB.ProgramCounter)

	// Decodificamos la instruccion

	w.WriteHeader(http.StatusOK)
}
