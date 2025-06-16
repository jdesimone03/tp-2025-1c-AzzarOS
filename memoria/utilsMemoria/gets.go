package utilsMemoria

import (
	"encoding/json"
	"log"
	"net/http"
)


func HandlerMostrarMetricas(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	metricasJSON, err := json.Marshal(Metricas)
	if err != nil {
		log.Println("Error al convertir las métricas a JSON", "error", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(metricasJSON)
	log.Println("Métricas enviadas")
}

func MostrarOcupadas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	ocupadasJSON, err := json.Marshal(Ocupadas)
	if err != nil {
		log.Println("Error al convertir las ocupadas a JSON", "error", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(ocupadasJSON)
	log.Println("Ocupadas enviadas")
}

func MostrarTablasDePaginas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	tablaJSON, err := json.Marshal(TablasDePaginas)
	if err != nil {
		log.Println("Error al convertir las tablas de páginas a JSON", "error", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(tablaJSON)
	log.Println("Tablas de Páginas enviadas")
}

func HandlerMostrarProcesoConInstrucciones(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	procesoConInstruccionesJSON, err := json.Marshal(Procesos)
	if err != nil {
		log.Println("Error al convertir las instrucciones a JSON", "error", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(procesoConInstruccionesJSON)
	log.Println("Lista de Procesos + PID enviada")
}