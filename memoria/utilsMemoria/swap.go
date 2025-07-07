package utilsMemoria

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"utils/logueador"
	"utils/structs"
)


func Tamanioframe() int {
	return Config.PageSize
}

func EscribirProcesoEsSwap(proceso structs.ProcesoEnSwap) {
	pathCorrecto := filepath.Base(Config.SwapfilePath)
	file, err := os.OpenFile(pathCorrecto, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logueador.Info("Error al abrir el archivo SWAP para escritura: %v", err)
		return
	}
	defer file.Close()

	dataJSON, error := json.Marshal(proceso)
	if error != nil {
		logueador.Info("Error al convertir el proceso a JSON")
		return
	}
	_, err = file.Write(append(dataJSON, '\n')) // Agrego salto de linea para que esten uno abajo del otro 
	if err != nil {
		logueador.Info("Error al escribir el proceso en el archivo SWAP: %v", err)
		return
	}

}

func CreacionArchivoSWAP() {
	pathCorrecto := filepath.Base(Config.SwapfilePath)
	file, err := os.Create(pathCorrecto)
	if err != nil {
	logueador.Info("Error al crear el archivo SWAP: %v", err)
	}
	logueador.Info("Archivo SWAP creado exitosamente en: %s", pathCorrecto)
	defer file.Close()
}

func BuscarPaginasDeProceso(pid uint) []string {
	var listaDePaginas []string
	
	for i := 0; i < CantidadDeFrames(); i++ {
		if Ocupadas[i] == int(pid) {
			leido, err := Read(pid, i * Tamanioframe(), Tamanioframe())	
			if err != nil {
				logueador.Info("Error al leer la página del proceso: %v", err)
				continue
			}
			listaDePaginas = append(listaDePaginas, leido)
			Ocupadas[i] = -1 // Marcar el frame como libre
		}
	}
	logueador.Info("Páginas encontradas para el proceso")
	return listaDePaginas
}

func BuscarProcesoEnSwap(pid uint) *structs.ProcesoEnSwap {
	pathCorrecto := filepath.Base(Config.SwapfilePath)
	file, err := os.Open(pathCorrecto)
	if err != nil {
		logueador.Info("Error al abrir el archivo SWAP: %v", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if err := scanner.Err(); err != nil {
		logueador.Info("Error al leer el archivo SWAP: %v", err)
	}
	var listaProcesos []structs.ProcesoEnSwap
	for scanner.Scan() {
		var proceso structs.ProcesoEnSwap
		err := json.Unmarshal(scanner.Bytes(), &proceso)
		if err != nil {
			logueador.Info("Error al deserializar el proceso desde el archivo SWAP: %v", err)
			continue // error que no tiene que frenarnos 
		}
		listaProcesos = append(listaProcesos, proceso)
	}

	procesoEncontrado, listaProcesos := ProcesoASacarDeSwap(listaProcesos, pid)
	if procesoEncontrado == nil {
		logueador.Info("No se encontró el proceso con PID %d en SWAP", pid)
		return nil
	}

	err = os.WriteFile(pathCorrecto, []byte{}, 0644) // Limpiar el archivo SWAP
	if err != nil {
		logueador.Info("Error al limpiar el archivo SWAP: %v", err)
		return nil
	}

	for i:= 0; i < len(listaProcesos); i++ {
		EscribirProcesoEsSwap(listaProcesos[i])
	} 
	return procesoEncontrado
}

func ProcesoASacarDeSwap(procesos []structs.ProcesoEnSwap, pid uint) (*structs.ProcesoEnSwap, []structs.ProcesoEnSwap) {
	// Buscar el proceso en la lista de procesos
	for i,proceso := range procesos {
		if proceso.PID == pid {
			encontrado := proceso
			logueador.Info("Proceso encontrado en SWAP: %+v", proceso)
			procesos = append(procesos[:i], procesos[i+1:]...) // Elimina el proceso encontrado de la lista
			return &encontrado, procesos // Retorna el proceso encontrado
		}
	}
	logueador.Info("Proceso con PID %d no encontrado en SWAP", pid)
	return nil, nil // Si no se encuentra, retorna nil
}
// Si hay espacio para inicializar => que entren las paginas del proceso en memoria principal
func SwapOutProceso(pid uint) {
	
	logueador.Info("Buscando proceso en SWAP")
	procesoEnSwap := BuscarProcesoEnSwap(pid)
	if procesoEnSwap == nil {
		logueador.Info("El proceso %d no se encuentra en el SWAP", pid)
		return
	}
	AcomodarProcesoEnMemoria(procesoEnSwap, pid)
}

func AcomodarProcesoEnMemoria(procesoEnSwap *structs.ProcesoEnSwap, pid uint) {
		// Volver a asignar las páginas al proceso en memoria principal
	for i:=0; i < len(procesoEnSwap.Paginas); i++ {
		frameLibre := PrimerFrameLibre(0)
		dirFisica := frameLibre * Tamanioframe()
		copy(EspacioUsuario[dirFisica:dirFisica + Tamanioframe()], []byte(procesoEnSwap.Paginas[i]))
		MarcarFrameOcupado(frameLibre, pid) 
	}
}

func SwapInProceso(pid uint) {
	paginas := BuscarPaginasDeProceso(pid)
	if len(paginas) == 0 {
		logueador.Info("No se encontraron páginas para el proceso %d", pid)
		return
	}
	procesoEnSwap := structs.ProcesoEnSwap{
		PID: pid,
		Paginas: paginas,
	}
	EscribirProcesoEsSwap(procesoEnSwap)
	logueador.Info("Proceso de PID: %d ha sido movido a SWAP", pid)
}
