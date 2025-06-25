package utilsMemoria

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"utils/logueador"
)

var pathCorrectoSwap string = filepath.Base(Config.SwapfilePath) // Asegura que el path sea correcto y no tenga problemas de directorios
var Tamanioframe = Config.PageSize


type ProcesoEnSwap struct {
	PID uint `json:"pid"` // Identificador del proceso
	Paginas []string `json:"paginas"` // Lista de páginas del proceso
}

func EscribirProcesoEsSwap(proceso ProcesoEnSwap) {
	file, err := os.OpenFile(pathCorrectoSwap, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error al abrir el archivo SWAP para escritura:", err)
		return
	}
	defer file.Close()

	dataJSON, error := json.Marshal(proceso)
	if error != nil {
		log.Println("Error al convertir el proceso a JSON:", error)
		return
	}
	_, err = file.Write(append(dataJSON, '\n')) // Agrego salto de linea para que esten uno abajo del otro 
	if err != nil {
		log.Println("Error al escribir el proceso en el archivo SWAP:", err)
		return
	}

}

func CreacionArchivoSWAP() {
	file, err := os.Create(pathCorrectoSwap)
	logueador.Info("Error al crear el archivo SWAP: %v", err)
	defer file.Close()
}

func BuscarPaginasDeProceso(pid uint) []string {
	var listaDePaginas []string
	
	for i := 0; i < CantidadDeFrames; i++ {
		if Ocupadas[uint(i)].PID == pid {
			leido, err := Read(pid, i * Tamanioframe, Tamanioframe)	
			if err != nil {
				log.Println("Error al leer la página del proceso:", err)
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
		log.Println("Error al abrir el archivo SWAP:", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var proceso ProcesoEnSwap
		err := json.Unmarshal(scanner.Bytes(), &proceso)
		if err != nil {
			log.Println("Error al deserializar el proceso desde el archivo SWAP:", err)
			continue // error que no tiene que frenarnos 
		}
		if proceso.PID == pid {
			log.Println("Proceso encontrado en SWAP:", proceso)
			return &proceso // Retorna el proceso encontrado
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error al leer el archivo SWAP:", err)
	}

	return nil
}

// Si hay espacio para inicializar => que entren las paginas del proceso en memoria principal
func SwapOutProceso(pid uint) {
	
	procesoEnSwap := BuscarProcesoEnSwap(pid)
	if procesoEnSwap == nil {
		log.Println("El proceso", pid, "no se encuentra en el SWAP")
		return
	}
	// Volver a asignar las páginas al proceso en memoria principal
	for i:=0; i < len(procesoEnSwap.Paginas); i++ {
		frameLibre := PrimerFrameLibre(0)
		dirFisica := frameLibre * Tamanioframe
		copy(EspacioUsuario[dirFisica:dirFisica + Tamanioframe], []byte(procesoEnSwap.Paginas[i]))
		MarcarFrameOcupado(uint(frameLibre), pid) 
		log.Println("Se le asigno el frame", frameLibre, "al proceso", pid,)
	}
}

func SwapInProceso(pid uint) {
	paginas := BuscarPaginasDeProceso(pid)
	if len(paginas) == 0 {
		log.Println("No se encontraron páginas para el proceso", pid)
		return
	}
	procesoEnSwap := ProcesoEnSwap{
		PID: pid,
		Paginas: paginas,
	}
	EscribirProcesoEsSwap(procesoEnSwap)
	log.Println("Proceso", pid, "ha sido movido al swap con las siguientes páginas:", paginas)
}
