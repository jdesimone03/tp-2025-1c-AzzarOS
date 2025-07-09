package utilsMemoria

import (
	"fmt"
	"utils/logueador"
	"utils/structs"
)

// -------------------------------- Tablas de Páginas --------------------------------

func InicializarEntrada(pid uint, frameAsignado int, entrada structs.EntradaDeTabla) structs.EntradaDeTabla {
	entrada.BitPresencia = false          // Inicialmente no esta en memoria
	entrada.BitModificado = false         // Inicialmente no se ha modificado
	entrada.NumeroDeFrame = frameAsignado // Inicialmente no tiene marco asignado
	MarcarFrameOcupado(frameAsignado, pid)
	return entrada
}

// -------------------------------- Manejo de Memoria --------------------------------

func HayFramesDisponibles(n int) bool {
	cant := 0
	for i := range Ocupadas {
		if Ocupadas[i] == -1 { // Si el frame está libre
			cant++
		}
		if cant >= n { // Si hay suficientes frames libres
			return true
		}

	}
	return false
}

func HayEspacioParaInicializar(tamanio int) bool {
	cantPaginas := CantidadDePaginasDeProceso(tamanio) // Cantidad de paginas que se necesitan
	if HayFramesDisponibles(cantPaginas) {             // Si hay suficientes frames libres
		return true
	} else {
		return false
	}
}

func CantidadDePaginasDeProceso(tamanio int) int {
	tamanioPagina := int(Config.PageSize)
	cantPaginas := (tamanio + tamanioPagina - 1) / tamanioPagina // Redondea hacia arriba la cantidad de páginas necesarias
	return cantPaginas
}

func Read(pid uint, direccion int, tamanio int) (string, error) {
	finDeLectura := direccion + tamanio

	if direccion < 0 || finDeLectura > Config.MemorySize {
		logueador.Error("Dirección fuera de rango %d - %d", direccion, finDeLectura)
		return "", nil // Retorna un error o un string vacío
	}

	if !ProcesoEnMemoria(pid) {
		logueador.Error("El proceso %d no está en memoria", pid)
		return "", fmt.Errorf("el proceso %d no está en memoria", pid)
	}

	inicio := direccion
	fin := direccion + tamanio - 1 // Verificar que la dirección de inicio y fin estén dentro del rango de memoria

	frameInicio := inicio / Config.PageSize
	frameFin := fin / Config.PageSize

	// Verificar que todos los frames involucrados pertenezcan al proceso
	for frame := frameInicio; frame <= frameFin; frame++ {
		if Ocupadas[frame] != int(pid) {
			logueador.Error("El proceso %d intenta escribir en el frame %d, que pertenece al PID %d", pid, frame, Ocupadas[frame])
			return "", fmt.Errorf("el proceso %d intenta escribir en el frame %d, que pertenece al PID %d", pid, frame, Ocupadas[frame])
		}
	}

	// Agrgar verificación por si se quiere leer otro frame que no le pertenece al proceso
	datosLeidos := make([]byte, tamanio)
	copy(datosLeidos, EspacioUsuario[direccion:finDeLectura]) // Copia los datos de la memoria principal a los datosLeidos

	return string(datosLeidos), nil // Retorna los datos leídos como string
}

func Write(pid uint, direccionFisica int, aEscribir string) error {

	if direccionFisica < 0 || direccionFisica >= Config.MemorySize {
		logueador.Error("Dirección fuera de rango: %d", direccionFisica)
		return fmt.Errorf("dirección fuera de rango: %d", direccionFisica)
	}

	inicio := direccionFisica
	fin := direccionFisica + len(aEscribir) - 1 // Verificar que la dirección de inicio y fin estén dentro del rango de memoria

	frameInicio := inicio / Config.PageSize
	frameFin := fin / Config.PageSize

	// Verificar que todos los frames involucrados pertenezcan al proceso
	for frame := frameInicio; frame <= frameFin; frame++ {
		if Ocupadas[frame] != int(pid) {
			logueador.Error("El proceso %d intenta escribir en el frame %d, que pertenece al PID %d", pid, frame, Ocupadas[frame])
			return fmt.Errorf("el proceso %d intenta escribir en el frame %d, que pertenece al PID %d", pid, frame, Ocupadas[frame])
		}
	}

	logueador.Info("Escribiendo en memoria")
	copy(EspacioUsuario[direccionFisica:], []byte(aEscribir))

	return nil
}

func ProcesoEnMemoria(pid uint) bool {
	for _, valor := range Ocupadas {
		if valor == int(pid) {
			return true
		}
	}
	return false
}
func CantidadDePaginas(pid uint) int {
	count := 0
	for _, frame := range Ocupadas {
		if Ocupadas[frame] == int(pid) {
			count++
		}
	}
	return count
}

func LiberarMemoria(pid uint) {
	delete(Procesos, pid) // Borramos el proceso de la tabla de procesos
	delete(TDPMultinivel, pid) // Borramos su tabla de páginas
	for i := range Ocupadas {
		if Ocupadas[i] == int(pid) {
			Ocupadas[i] = -1 // Marca el frame como libre
			logueador.Info("Liberando frame %d del proceso %d", i, pid)
		}
	}
}

func MarcarFrameOcupado(frame int, pid uint) {
	Ocupadas[frame] = int(pid)
}

func InicializarOcupadas() {
	tam := Config.MemorySize / Config.PageSize // Cantidad de frames
	Ocupadas = make([]int, tam)                // Inicializa la lista de frames ocupados con -1
	for i := range Config.MemorySize/Config.PageSize {
		Ocupadas[i] = -1
	}
}