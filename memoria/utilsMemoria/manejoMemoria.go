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
    for _, frame := range Ocupadas {
        if !frame.EstaOcupado {
            cant++
            if cant >= n {
                return true
            }
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
	cantPaginas := (tamanio / tamanioPagina) // Cantidad de paginas que se necesitan
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
		if Ocupadas[frame].PID != pid {
			logueador.Error("El proceso %d intenta escribir en el frame %d, que pertenece al PID %d", pid, frame, Ocupadas[frame].PID)
			return "", fmt.Errorf("el proceso %d intenta escribir en el frame %d, que pertenece al PID %d", pid, frame, Ocupadas[frame].PID)
		}
	}

	// Agrgar verificación por si se quiere leer otro frame que no le pertenece al proceso
	datosLeidos := make([]byte, tamanio)
	copy(datosLeidos, EspacioUsuario[direccion:finDeLectura]) // Copia los datos de la memoria principal a los datosLeidos

	return string(datosLeidos), nil // Retorna los datos leídos como string
}

func Write(pid uint, direccion int, aEscribir string) (error) {

	if direccion < 0 || direccion >= Config.MemorySize {
		logueador.Error("Dirección fuera de rango: %d", direccion)
		return fmt.Errorf("dirección fuera de rango: %d", direccion)
	}

	inicio := direccion
	fin := direccion + len(aEscribir) - 1 // Verificar que la dirección de inicio y fin estén dentro del rango de memoria

	frameInicio := inicio / Config.PageSize
	frameFin := fin / Config.PageSize

	// Verificar que todos los frames involucrados pertenezcan al proceso
	for frame := frameInicio; frame <= frameFin; frame++ {
		if Ocupadas[frame].PID != pid {
			logueador.Error("El proceso %d intenta escribir en el frame %d, que pertenece al PID %d", pid, frame, Ocupadas[frame].PID)
			return fmt.Errorf("el proceso %d intenta escribir en el frame %d, que pertenece al PID %d", pid, frame, Ocupadas[frame].PID)
		}
	}

	logueador.Info("Escribiendo en memoria")
	copy(EspacioUsuario[direccion:], []byte(aEscribir))
	return nil 
}

func ProcesoEnMemoria(pid uint) bool {
	for _, frame := range Ocupadas {
		if frame.PID == pid{
			return true // Si hay al menos un frame ocupado por el PID, el proceso está en memoria
		}
	}
	return false // Si no se encuentra ningún frame ocupado por el PID, el proceso no está en memoria
}

func CantidadDePaginas(pid uint) int {
	count := 0
	for _, frame := range Ocupadas {
		if frame.PID == pid && frame.EstaOcupado {
			count++
		}
	}
	return count
}

func LiberarMemoria(pid uint) {
	for i := 0; i < CantidadDePaginas(pid); i++ {
		frame := Ocupadas[i]
		if frame.PID == pid {
			frame.EstaOcupado = false
			frame.PID = 0 // TODO
			Ocupadas[i] = frame
			logueador.Info("Liberando frame %d del proceso %d", i, pid)
		}
	}
}

func MarcarFrameOcupado(frame int, pid uint) {
	info := Ocupadas[frame]
	info.EstaOcupado = true // Marca el frame como ocupado
	info.PID = pid          // Asigna el PID del proceso que ocupa el frame
	Ocupadas[frame] = info  // Actualiza el mapa con la información del frame
}

func InicializarOcupadas() {
	Ocupadas = make(map[int]structs.FrameInfo)
	for i := 0; i < Config.MemorySize/Config.PageSize; i++ {
		Ocupadas[i] = structs.FrameInfo{
			EstaOcupado: false,
			PID:         0,
		}
	}
}