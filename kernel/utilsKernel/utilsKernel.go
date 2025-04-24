package utilsKernel

import (
	"fmt"
	"log/slog"
	"net/http"
	"utils"
	"utils/config"
	"utils/structs"
)

// variables Config
var Config = config.CargarConfiguracion[config.ConfigKernel]("config.json")

// scheduler_algorithm: LARGO plazo
// ready_ingress_algorithm: CORTO plazo

var Interfaces = make(map[string]structs.Interfaz)

// Handlers de endpoints
func RecibirInterfaz(w http.ResponseWriter, r *http.Request) {
	interfaz, err := utils.DecodificarMensaje[structs.PeticionIO](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	addInterfaz(interfaz.Nombre, interfaz.Interfaz)
	slog.Info(fmt.Sprintf("Me llego la siguiente interfaz: %+v", interfaz))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func HandleSyscall(w http.ResponseWriter, r *http.Request) {
	// Despues le hacemos un case switch para cada syscall diferente
	peticion, err := utils.DecodificarMensaje[structs.PeticionKernel](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}
	ifazElegida := Interfaces[peticion.NombreIfaz]
	peticion.SuspensionTime = Config.SuspensionTime
	utils.EnviarMensaje(ifazElegida.IP, ifazElegida.Puerto, "peticionKernel", peticion)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func addInterfaz(nombre string, ifaz structs.Interfaz) {
	Interfaces[nombre] = ifaz
}
