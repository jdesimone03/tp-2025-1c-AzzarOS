package utilsMemoria

import (
	"encoding/json"
	"utils/logueador"
	"net/http"
	"bufio"
	"os"
	"path/filepath"
	"utils/structs"
	"fmt"
)

func MostrarMemoria(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	const ancho = 32 // Cuántos bytes mostrar por línea
	for i := 0; i < len(EspacioUsuario); i += ancho {
		fmt.Fprintf(w, "%04X  ", i)

		// Mostrar bytes como caracteres legibles
		for j := 0; j < ancho && i+j < len(EspacioUsuario); j++ {
			b := EspacioUsuario[i+j]
			if b >= 32 && b <= 126 {
				fmt.Fprintf(w, "%c", b)
			} else {
				fmt.Fprintf(w, ".")
			}
		}
		fmt.Fprintln(w)
	}

	logueador.Info("Memoria enviada en formato texto legible")
}

func HandlerMostrarSWAP(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	
	pathCorrecto := filepath.Base(Config.SwapfilePath)
	file, err := os.Open(pathCorrecto)
	if err != nil {
		logueador.Info("Error al abrir el archivo SWAP: %s", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var procesos []structs.ProcesoEnSwap

	for scanner.Scan() {
		var proceso structs.ProcesoEnSwap
		err := json.Unmarshal(scanner.Bytes(), &proceso)
		if err != nil {
			logueador.Info("Error al decodificar el proceso: %s", err)
			continue
		}
		procesos = append(procesos, proceso)
	}

	if err := scanner.Err(); err != nil {
		logueador.Info("Error al leer el archivo SWAP: %s", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	procesosJSON, err := json.Marshal(procesos)
	if err != nil {
		logueador.Info("Error al convertir los procesos a JSON: %s", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(procesosJSON)
	logueador.Info("Procesos en SWAP enviados")

}



func HandlerMostrarMetricas(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	metricasJSON, err := json.Marshal(Metricas)
	if err != nil {
		logueador.Info("Error al convertir las métricas a JSON %s", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(metricasJSON)
	logueador.Info("Métricas enviadas")
}

func MostrarOcupadas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	ocupadasJSON, err := json.Marshal(Ocupadas)
	if err != nil {
		logueador.Info("Error al convertir las ocupadas a JSON %s", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(ocupadasJSON)
	logueador.Info("Ocupadas enviadas")
}


func HandlerMostrarProcesoConInstrucciones(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	procesoConInstruccionesJSON, err := json.Marshal(Procesos)
	if err != nil {
		logueador.Info("Error al convertir las instrucciones a JSON %s", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(procesoConInstruccionesJSON)
	logueador.Info("Lista de Procesos + PID enviada")
}