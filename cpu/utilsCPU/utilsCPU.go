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
		return
	}

	// Solicitar a memoria la siguiente instruccion para la ejecución
	// InstruccionCodificada = FETCH(PCB.ProgramCounter)

	// Decodificamos la instruccion

	w.WriteHeader(http.StatusOK)
}

func FetchAndDecode(peticion structs.PeticionMemoria) any{
	instruccion := utils.EnviarMensaje(Config.IPMemory,Config.PortMemory,"fetch",peticion)
	instruccionDecodificada := utils.Decode(instruccion)
	return instruccionDecodificada
}

//AUMENTAR PC NO SE EN QUE MOMENTO SE HACE QUIERO VER LA TEORIA ANTES DE IMPLEMENTARLO

func Execute(instruccion any){
	switch instruccion.(type){
		case structs.NoopInstruction:
			//hace nada
		case structs.WriteInstruction:
			//hace lo que tenga quer hacer
		case structs.ReadInstruction:
			//hace lo que tenga quer hacer
		case structs.GotoInstruction:
			//hace lo que tenga quer hacer
		case structs.IOInstruction:
			//hace lo que tenga quer hacer
		case structs.InitProcInstruction:
			//hace lo que tenga quer hacer
		case structs.DumpMemoryInstruction:
			//hace lo que tenga quer hacer
		case structs.ExitInstruction:
			//hace lo que tenga quer hacer
		default:
			//si llega algo inesperado
			slog.Error("Tipo de instrucción desconocido")
	}
}
