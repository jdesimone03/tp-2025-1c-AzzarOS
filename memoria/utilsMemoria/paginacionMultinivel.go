package utilsMemoria

import (
	"utils/structs"
)


func FrameLibre(numero int) bool {
	return Ocupadas[numero] != -1 
}

func PrimerFrameLibre(arranque int) int { // arranque => desde cual frame arranco a buscar
	CantidadDeFrames := len(Ocupadas)
    for i := arranque; i < CantidadDeFrames; i++ {
        if Ocupadas[i] == -1 { // Si el frame esta libre
            return i
        }
    }
    return -1
}

func PrimerFrameLibreSinLogs(arranque int) int {
	CantidadDeFrames := len(Ocupadas)
    for i := arranque; i < CantidadDeFrames; i++ {
        if Ocupadas[i] == -1 {
            return i
        }
    }
    return -1
}


func CreaTablaJerarquica(pid uint, nivelesRestantes int, paginasRestantes *int) *structs.Tabla {

	tabla := &structs.Tabla{}

	if nivelesRestantes == 1 {
		// Nivel hoja: asignar solo las páginas necesarias
		for i := 0; i < Config.EntriesPerPage; i++ {
			if *paginasRestantes > 0 {
				frameLibre := PrimerFrameLibreSinLogs(i) // Busca el primer frame libre 
 				MarcarFrameOcupado(frameLibre, pid) // Lo marca como ocupado para el PID
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
	paginasRestantes := CantidadDePaginasDeProceso(tamanio)
	tabla := CreaTablaJerarquica(pid, Config.NumberOfLevels, &paginasRestantes) // Crea una tabla jerárquica para el PID
	TDPMultinivel[pid] = tabla // Asigna la tabla al mapa TDP
}

