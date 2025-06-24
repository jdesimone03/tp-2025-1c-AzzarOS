package utilsCPU

import (
	"fmt"
	"net/http"
	"utils"
	"utils/logueador"
	"utils/structs"
)

// ---------------------------- Handlers ----------------------------//
func RecibirInterrupcion(w http.ResponseWriter, r *http.Request) {

	// Log obligatorio 2/11
	logueador.InterrupcionRecibida()
	InterruptFlag = true

	w.WriteHeader(http.StatusOK)
}

func RecibirEjecucion(w http.ResponseWriter, r *http.Request) {
	ejecucion, err := utils.DecodificarMensaje[structs.EjecucionCPU](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje (%v)", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf("http://%s:%d/proximaInstruccion", Config.IPMemory, Config.PortMemory)
	logueador.Info("Enviando mensaje a memoria: %s", url)




	// chEjecucion <- *ejecucion

	w.WriteHeader(http.StatusOK)
}
