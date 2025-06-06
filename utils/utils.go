package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"utils/structs"
)

func EnviarMensaje(ip string, puerto int, endpoint string, mensaje any) string {
	body, err := json.Marshal(mensaje)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo codificar el mensaje (%v)", err))
		return ""
	}

	url := fmt.Sprintf("http://%s:%d/%s", ip, puerto, endpoint)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo enviar mensaje a %s:%d/%s (%v)", ip, puerto, endpoint, err))
		return ""
	}
	defer resp.Body.Close()

	var resData structs.Respuesta
	respuesta, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	err = json.Unmarshal(respuesta, &resData)
	if err != nil {
		return ""
	}

	// log.Printf("respuesta del servidor: %s", resp.Status)
	slog.Info(fmt.Sprintf("Respuesta de %s:%d/%s %v", ip, puerto, endpoint, resData))
	return resData.Mensaje
}

// Función genérica para decodificar un mensaje del body
func DecodificarMensaje[T any](r *http.Request) (*T, error) {
	var mensaje T
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&mensaje)
	if err != nil {
		return nil, err
	}
	return &mensaje, nil
}

func DecodificarSyscall(r *http.Request, tipoSyscall string) (*structs.SyscallInstruction, error) {
    var raw struct {
        PID         uint            `json:"pid"`
        Instruccion json.RawMessage `json:"instruccion"`
    }

    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&raw)
    if err != nil {
        return nil, err
    }

	 // Log para verificar el contenido de Instruccion
	 slog.Info(fmt.Sprintf("Instrucción recibida (raw): %s", string(raw.Instruccion)))

    var instruccion structs.Syscall
    switch tipoSyscall {
    case "IO":
        var ioInst structs.IOInstruction
        if err := json.Unmarshal(raw.Instruccion, &ioInst); err != nil {
            return nil, fmt.Errorf("error al decodificar IOInstruction: %w. JSON: %s", err, string(raw.Instruccion))
        }
        instruccion = &ioInst
    case "INIT_PROC":
        var initProcInst structs.InitProcInstruction
        if err := json.Unmarshal(raw.Instruccion, &initProcInst); err != nil {
            return nil, fmt.Errorf("error al decodificar InitProcInstruction: %w. JSON: %s", err, string(raw.Instruccion))
        }
        instruccion = &initProcInst
    case "DUMP_MEMORY":
        var dumpMemInst structs.DumpMemoryInstruction
        if err := json.Unmarshal(raw.Instruccion, &dumpMemInst); err != nil {
            return nil, fmt.Errorf("error al decodificar DumpMemoryInstruction: %w. JSON: %s", err, string(raw.Instruccion))
        }
        instruccion = &dumpMemInst
    case "EXIT":
        var exitInst structs.ExitInstruction
        if err := json.Unmarshal(raw.Instruccion, &exitInst); err != nil {
            return nil, fmt.Errorf("error al decodificar ExitInstruction: %w. JSON: %s", err, string(raw.Instruccion))
        }
        instruccion= &exitInst
    default:
        return nil, fmt.Errorf("tipo de instrucción desconocido")
    }

    return &structs.SyscallInstruction{
        PID:         raw.PID,
        Instruccion: instruccion,
    }, nil
}

func IniciarServidor(puerto int) error {
	slog.Info(fmt.Sprintf("Inicializando servidor en el puerto %d",puerto))
	err := http.ListenAndServe(fmt.Sprintf(":%d", puerto), nil)
	if err != nil {
		panic(err)
	}
	return err
}
