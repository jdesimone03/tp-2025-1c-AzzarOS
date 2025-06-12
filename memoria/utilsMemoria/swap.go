package utilsMemoria

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"utils/logueador"
)

var pathCorrectoSwap string = filepath.Base(Config.SwapfilePath) // Asegura que el path sea correcto y no tenga problemas de directorios

type ProcesoEnSwap struct {
	PID     uint     `json:"pid"`     // Identificador del proceso
	Paginas []string `json:"paginas"` // Lista de páginas del proceso
}

func EscribirProcesoEsSwap(proceso ProcesoEnSwap) {
	file, err := os.OpenFile(pathCorrectoSwap, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logueador.Error("Error al abrir el archivo SWAP para escritura: %e", err)
		return
	}
	defer file.Close()

	dataJSON, err := json.Marshal(proceso)
	if err != nil {
		logueador.Error("Error al convertir el proceso a JSON: %e", err)
		return
	}
	_, err = file.Write(append(dataJSON, '\n')) // Agrego salto de linea para que esten uno abajo del otro
	if err != nil {
		logueador.Error("Error al escribir el proceso en el archivo SWAP: %e", err)
		return
	}

}

func check(mensaje string, e error) {
	if e != nil {
		logueador.Error("%s: %e", mensaje, e)
	}
}

func CreacionArchivoSWAP() {
	file, err := os.Create(pathCorrectoSwap)
	defer file.Close()
	check("Error al crear el archivo SWAP", err)
}

func BuscarPaginasDeProceso(pid uint) []string {
	var listaDePaginas []string

	for i := 0; i < CantidadDeFrames; i++ {
		if Ocupadas[uint(i)].PID == pid {
			leido, err := Read(pid, i*Config.PageSize, Config.PageSize)
			if err != nil {
				logueador.Error("Error al leer la página del proceso: %e", err)
				continue
			}
			listaDePaginas = append(listaDePaginas, leido)
			frame := Ocupadas[uint(i)]
			frame.EstaOcupado = false // Marcar el frame como libre
			Ocupadas[uint(i)] = frame
		}
	}
	return listaDePaginas
}

func BuscarProcesoEnSwap(pid uint) *ProcesoEnSwap {
	file, err := os.Open(pathCorrectoSwap)
	if err != nil {
		logueador.Error("Error al abrir el archivo SWAP: %e", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var proceso ProcesoEnSwap
		err := json.Unmarshal(scanner.Bytes(), &proceso)
		if err != nil {
			logueador.Error("Error al deserializar el proceso desde el archivo SWAP: %e", err)
			continue // error que no tiene que frenarnos
		}
		if proceso.PID == pid {
			logueador.Info("Proceso encontrado en SWAP: %+v", proceso)
			return &proceso // Retorna el proceso encontrado
		}
	}

	if err := scanner.Err(); err != nil {
		logueador.Error("Error al leer el archivo SWAP: %e", err)
	}

	return nil
}

// Si hay espacio para inicializar => que entren las paginas del proceso en memoria principal
func SwapOutProceso(pid uint) {

	procesoEnSwap := BuscarProcesoEnSwap(pid)
	if procesoEnSwap == nil {
		logueador.Error("El proceso %d no se encuentra en el SWAP", pid)
		return
	}
	// Volver a asignar las páginas al proceso en memoria principal
	for i := 0; i < len(procesoEnSwap.Paginas); i++ {
		frameLibre := PrimerFrameLibre(0)
		dirFisica := frameLibre * Config.PageSize
		copy(EspacioUsuario[dirFisica:dirFisica+Config.PageSize], []byte(procesoEnSwap.Paginas[i]))
		MarcarFrameOcupado(uint(frameLibre), pid)
		logueador.Info("Se le asigno el frame %d al proceso %d", frameLibre, pid)
	}
}

func SwapInProceso(pid uint) {
	paginas := BuscarPaginasDeProceso(pid)
	if len(paginas) == 0 {
		logueador.Error("No se encontraron páginas para el proceso %d", pid)
		return
	}
	procesoEnSwap := ProcesoEnSwap{
		PID:     pid,
		Paginas: paginas,
	}
	EscribirProcesoEsSwap(procesoEnSwap)
	logueador.Info("Proceso %d ha sido movido al swap con las siguientes páginas: %v", pid, paginas)
}
