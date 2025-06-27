package utilsMemoria

import (
	"utils/logueador"
	"utils/structs"
)



// -------------------------------- Tablas de Páginas --------------------------------

func InicializarEntrada(pid uint, frameAsignado int, entrada structs.EntradaDeTabla) structs.EntradaDeTabla {
	entrada.BitPresencia = false          // Inicialmente no esta en memoria
	entrada.BitModificado = false         // Inicialmente no se ha modificado
	entrada.NumeroDeFrame = frameAsignado // Inicialmente no tiene marco asignado
	MarcarFrameOcupado(uint(frameAsignado), pid)
	return entrada
}

// -------------------------------- Manejo de Memoria --------------------------------


func HayFramesDisponibles(n int) bool {
	i := 0
	cant := 0
	for cant < n {
		frameLibre := PrimerFrameLibre(uint(i)) // Busco el primer frame libre a partir del indice i
		if frameLibre == -1 {                   // Si no hay mas frames libres, salgo
			logueador.Warn("No hay suficientes frames libres")
			return false
		}
		cant++ // Si hay un frame libre, aumento la cantidad de frames libres encontrados
		i++
	}
	return true
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

	// Agrgar verificación por si se quiere leer otro frame que no le pertenece al proceso
	datosLeidos := make([]byte, tamanio)
	copy(datosLeidos, EspacioUsuario[direccion:finDeLectura]) // Copia los datos de la memoria principal a los datosLeidos

	return string(datosLeidos), nil // Retorna los datos leídos como string
}

func Write(pid uint, direccion int, aEscribir string) {

	if direccion < 0 || direccion >= Config.MemorySize {
		logueador.Error("Dirección fuera de rango: %d", direccion)
		return
	}

	// paginasDeProceso := CantidadDePaginas(pid) // Obtiene la cantidad de paginas del proceso
	// if len(memoriaPrincipal[direccion:]) > paginasDeProceso * Tamanioframe {
	// 	logueador.Info("Los bytes a escribir exceden el tamaño de todas las páginas del proceso")
	// 	return
	// }

	frameAEscribir := int(direccion / Config.PageSize)
	if Ocupadas[uint(frameAEscribir)].PID != pid {
		logueador.Error("El proceso %d quiere escribir en el frame %d que no le pertenece", pid, frameAEscribir)
		return
	}

	logueador.Info("Escribiendo en memoria")
	copy(EspacioUsuario[direccion:], []byte(aEscribir))

	// tablasDePaginas := ObtenerTablaDePaginas(pid) // Obtengo la tabla de paginas del PID
	// tablasDePaginas.Entradas[frameAEscribir].BitModificado = true // Marca la entrada como presente en memoria

	IncrementarMetricaEn(pid, "Escrituras")
	logueador.Info("Escritura exitosa en memoria")
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
		frame := Ocupadas[uint(i)]
		if frame.PID == pid {
			frame.EstaOcupado = false
			frame.PID = 0 // TODO
			Ocupadas[uint(i)] = frame
			logueador.Info("Liberando frame %d del proceso %d", i, pid)
		}
	}
}

func MarcarFrameOcupado(frame uint, pid uint) {
	info := Ocupadas[frame]
	info.EstaOcupado = true // Marca el frame como ocupado
	info.PID = pid          // Asigna el PID del proceso que ocupa el frame
	Ocupadas[frame] = info  // Actualiza el mapa con la información del frame
}

func InicializarOcupadas() {
	Ocupadas = make(map[uint]structs.FrameInfo)
	for i := uint(0); i < uint(Config.MemorySize/Config.PageSize); i++ {
		Ocupadas[i] = structs.FrameInfo{
			EstaOcupado: false,
			PID:         0,
		}
	}
}