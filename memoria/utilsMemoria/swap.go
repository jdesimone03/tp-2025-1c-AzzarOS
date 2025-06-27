package utilsMemoria

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"utils/logueador"
)


func Tamanioframe() int {
	return Config.PageSize
}
type ProcesoEnSwap struct {
	PID uint `json:"pid"` // Identificador del proceso
	Paginas []string `json:"paginas"` // Lista de páginas del proceso
}

func EscribirProcesoEsSwap(proceso ProcesoEnSwap) {
	pathCorrecto := filepath.Base(Config.SwapfilePath)
	file, err := os.OpenFile(pathCorrecto, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logueador.Info("Error al abrir el archivo SWAP para escritura:", err)
		return
	}
	defer file.Close()

	dataJSON, error := json.Marshal(proceso)
	if error != nil {
		logueador.Info("Error al convertir el proceso a JSON:", error)
		return
	}
	_, err = file.Write(append(dataJSON, '\n')) // Agrego salto de linea para que esten uno abajo del otro 
	if err != nil {
		logueador.Info("Error al escribir el proceso en el archivo SWAP:", err)
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
		if Ocupadas[uint(i)].PID == pid {
			leido, err := Read(pid, i * Tamanioframe(), Tamanioframe())	
			if err != nil {
				logueador.Info("Error al leer la página del proceso:", err)
				continue
			}
			listaDePaginas = append(listaDePaginas, leido)
			frame := Ocupadas[uint(i)]
			frame.EstaOcupado = false // Marcar el frame como libre
			frame.PID = 0 // Limpiar el PID del frame
			Ocupadas[uint(i)] = frame
		}
	}
	logueador.Info("Páginas encontradas para el proceso" )
	return listaDePaginas
}

func BuscarProcesoEnSwap(pid uint) *ProcesoEnSwap {
	pathCorrecto := filepath.Base(Config.SwapfilePath)
	file, err := os.Open(pathCorrecto)
	if err != nil {
		logueador.Info("Error al abrir el archivo SWAP: %d", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var proceso ProcesoEnSwap
		err := json.Unmarshal(scanner.Bytes(), &proceso)
		if err != nil {
			logueador.Info("Error al deserializar el proceso desde el archivo SWAP:", err)
			continue // error que no tiene que frenarnos 
		}
		if proceso.PID == pid {
			logueador.Info("Proceso encontrado en SWAP:", proceso)
			return &proceso // Retorna el proceso encontrado
		}
	}

	if err := scanner.Err(); err != nil {
		logueador.Info("Error al leer el archivo SWAP:", err)
	}

	return nil
}

// Si hay espacio para inicializar => que entren las paginas del proceso en memoria principal
func SwapOutProceso(pid uint) {
	
	logueador.Info("Buscando proceso en SWAP")
	procesoEnSwap := BuscarProcesoEnSwap(pid)
	if procesoEnSwap == nil {
		logueador.Info("El proceso", pid, "no se encuentra en el SWAP")
		return
	}
	// Volver a asignar las páginas al proceso en memoria principal
	for i:=0; i < len(procesoEnSwap.Paginas); i++ {
		frameLibre := PrimerFrameLibre(0)
		dirFisica := frameLibre * Tamanioframe()
		copy(EspacioUsuario[dirFisica:dirFisica + Tamanioframe()], []byte(procesoEnSwap.Paginas[i]))
		MarcarFrameOcupado(uint(frameLibre), pid) 
	}
}

func SwapInProceso(pid uint) {
	paginas := BuscarPaginasDeProceso(pid)
	if len(paginas) == 0 {
		logueador.Info("No se encontraron páginas para el proceso", pid)
		return
	}
	procesoEnSwap := ProcesoEnSwap{
		PID: pid,
		Paginas: paginas,
	}
	EscribirProcesoEsSwap(procesoEnSwap)
	logueador.Info("Proceso de PID: %d ha sido movido a SWAP", pid)
}
