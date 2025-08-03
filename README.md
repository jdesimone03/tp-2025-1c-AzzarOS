# TP Sistemas Operativos 1C2025 "AzzarOS"
_"Gopher se cree que es ejemplo..."_
* [Consigna](https://docs.google.com/document/d/1zoFRoBn9QAfYSr0tITsL3PD6DtPzO2sq9AtvE8NGrkc/edit?usp=sharing)
* [Documento de pruebas](https://docs.google.com/document/d/13XPliZvUBtYjaRfuVUGHWbYX8LBs8s3TDdaDa9MFr_I/edit?usp=sharing)

## Integrantes
* [@jdesimone03](https://github.com/jdesimone03)
* [@SuperGriditOS](https://github.com/SuperGriditOS)
* [@santirondini](https://github.com/santirondini)
* [@GordoNobli](https://github.com/GordoNobli)

## Configuración
Ejecute el script de setup en `utils/scripts/setup.sh` y cambie las IPs de cada módulo. Opcionalmente puede cambiar los puertos, los niveles de logging y las rutas donde se crearán los archivos que guarden estructuras de memoria.
### Configuración manual
1. Copie el archivo `.env.example` y pongale de nombre `.env`.
2. Edite el archivo `.env` con toda la configuración necesaria.
3. Es posible que tenga que cambiar los permisos de ejecución para todos los scripts. Ejecute `chmod +x <script>` para solucionarlo.

## Testeo
El orden de inicialización es Memoria/Kernel → CPU/IO. Es importante seguir este orden para que se inicializen bien todas las estructuras.

Para ejecutar un módulo con su test correspondiente, simplemente ejecute el siguiente comando en el módulo que quiera iniciar.
```
./run_tests.sh <numero de test> <args...>
```
Ejecutando el comando sin argumentos muestra todas las opciones para ese módulo.

### Testeo con VSCode
Si abre el repositorio en Visual Studio Code, podrá realizar los tests con el debugger. Estos se pueden configurar en el archivo `.vscode/launch.json`.
