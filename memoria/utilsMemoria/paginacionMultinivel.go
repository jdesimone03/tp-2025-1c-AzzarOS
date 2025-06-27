package utilsMemoria

import (
	"utils/logueador"
	"utils/structs"
)


func FrameLibre(numero uint) bool {
	return !Ocupadas[numero].EstaOcupado
}

func PrimerFrameLibre(arranque uint) int { // arranque => desde cual frame arranco a buscar
	CantidadDeFrames := uint(len(Ocupadas)) // Cantidad de frames que hay en memoria
	for i := arranque; i < uint(CantidadDeFrames); i++ {
		if FrameLibre(i) {
			logueador.Info("Frame libre encontrado - NRO Frame %d", i)
			return int(i)
		}
	}
	logueador.Warn("No se encontraron frames libres")
	return -1 // Si no encuentra un frame libre => memoria llena => devuelvo -1
}

func PrimerFrameLibreSinLogs(arranque uint) int { // arranque => desde cual frame arranco a buscar
	CantidadDeFrames := uint(len(Ocupadas)) // Cantidad de frames que hay en memoria
	for i := arranque; i < uint(CantidadDeFrames); i++ {
		if FrameLibre(i) {
			return int(i)
		}
	}
	logueador.Warn("No se encontraron frames libres")
	return -1 // Si no encuentra un frame libre => memoria llena => devuelvo -1
}


func CreaTablaJerarquica(pid uint, nivelesRestantes int, paginasRestantes *int) *structs.Tabla {

	tabla := &structs.Tabla{}

	if nivelesRestantes == 1 {
		// Nivel hoja: asignar solo las páginas necesarias
		for i := 0; i < Config.EntriesPerPage; i++ {
			if *paginasRestantes > 0 {
				frameLibre := PrimerFrameLibreSinLogs(uint(i)) // Busca el primer frame libre 
 				MarcarFrameOcupado(uint(frameLibre), pid) // Lo marca como ocupado para el PID
				tabla.Valores = append(tabla.Valores, frameLibre)
				*paginasRestantes--
			} else {
				tabla.Valores = append(tabla.Valores, -1) // El valor -1 respresenta que no esta asignado 
			}
		}
	} else {
		// Nivel intermedio: crear subtablas recursivamente
		for i := 0; i < Config.EntriesPerPage; i++ {
			subtabla := CreaTablaJerarquica(pid, nivelesRestantes - 1,paginasRestantes)
			tabla.Punteros = append(tabla.Punteros, subtabla)
		}
	}
	return tabla
}

func CrearTablaDePaginas(pid uint, tamanio int) {
	logueador.Info("Creando tabla de paginas para el PID %d" , pid)
	paginasRestantes := CantidadDePaginasDeProceso(tamanio)
	tabla := CreaTablaJerarquica(pid, Config.NumberOfLevels, &paginasRestantes) // Crea una tabla jerárquica para el PID
	TDPMultinivel[pid] = tabla // Asigna la tabla al mapa TDP
}

