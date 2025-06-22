package utilsCPU

import (
	"utils/config"
	"utils/structs"
)

var Config config.ConfigCPU
var InterruptFlag bool
var chEjecucion = make(chan structs.EjecucionCPU)

func init() {
	go func() {
		for {
			// channel que avisa que va a ejecutar
			ctxEjecucion := <-chEjecucion
			Ejecucion(ctxEjecucion)
		}
	}()
}