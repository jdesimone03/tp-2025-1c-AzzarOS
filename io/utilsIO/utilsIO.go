package utilsIO

import (
	"fmt"
	"log/slog"
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
var hayEjecucion = make(chan bool)

// Manejo de señales
func init() {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			<-hayEjecucion
			Ejecucion(Ejecutando)
		}
	}()

	go func() {
		sig := <-sigs
		slog.Warn(fmt.Sprintf("Señal recibida: %v. Notificando al Kernel.", sig))
		utils.EnviarMensaje(Config.IPKernel, Config.PortKernel, "io-disconnect", NombreInterfaz)
		slog.Info("Notificación de desconexión enviada al Kernel.")
		os.Exit(0)
	}()
}

func RecibirEjecucionIO(w http.ResponseWriter, r *http.Request) {
	ejecucion, err := utils.DecodificarMensaje[structs.EjecucionIO](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	Ejecutando = *ejecucion
	hayEjecucion <- true

	w.WriteHeader(http.StatusOK)
}

func Ejecucion(ctx structs.EjecucionIO) {
	// Log obligatorio 1/2
	logueador.InicioIO(ctx.PID, ctx.TiempoMs)

	time.Sleep(time.Duration(ctx.TiempoMs) * time.Millisecond)

	// Log obligatorio 2/2
	logueador.FinalizacionIO(ctx.PID)

	utils.EnviarMensaje(Config.IPKernel, Config.PortKernel, "io-end", NombreInterfaz)

	hayEjecucion <- false
}