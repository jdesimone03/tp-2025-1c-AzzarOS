package utilsIO

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"utils"
	"utils/config"
	"utils/logueador"
	"utils/structs"
)

var Config config.ConfigIO
var NombreInterfaz string
var Ejecutando structs.EjecucionIO
var chEjecucion = make(chan structs.EjecucionIO, 1)

// Manejo de señales
func init() {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			exec := <-chEjecucion
			Ejecucion(exec)
		}
	}()
 
	go func() {
		sig := <-sigs
		logueador.Warn("Señal recibida: %v. Notificando al Kernel.", sig)
		utils.EnviarMensaje(Config.IPKernel, Config.PortKernel, "io-disconnect", NombreInterfaz)
		logueador.Info("Notificación de desconexión enviada al Kernel.")
		os.Exit(0)
	}()
}

func RecibirEjecucionIO(w http.ResponseWriter, r *http.Request) {
	ejecucion, err := utils.DecodificarMensaje[structs.EjecucionIO](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje (%v)", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	select {
    case chEjecucion <- *ejecucion:
        w.WriteHeader(http.StatusOK)
    case <-time.After(5 * time.Second): // Timeout de 5 segundos
        logueador.Error("Timeout al enviar ejecución al canal")
        w.WriteHeader(http.StatusInternalServerError)
    }

	//w.WriteHeader(http.StatusOK)
}

func Ejecucion(ctx structs.EjecucionIO) {
	// Log obligatorio 1/2
	logueador.InicioIO(ctx.PID, ctx.TiempoMs)

	time.Sleep(time.Duration(ctx.TiempoMs) * time.Millisecond)

	// Log obligatorio 2/2
	logueador.FinalizacionIO(ctx.PID)

	utils.EnviarMensaje(Config.IPKernel, Config.PortKernel, "io-end", NombreInterfaz)
}