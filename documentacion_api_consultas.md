================================================
  DOCUMENTACIÓN DE LA API - apiconsulta.sehuacho.com
================================================

1. URL BASE Y AUTENTICACIÓN
-----------------------------
Base URL: https://apiconsulta.sehuacho.com/api

Todas las peticiones deben incluir el siguiente header:
  x-api-key: tu_api_key_aqui


2. ENDPOINTS DISPONIBLES
--------------------------
[GET] /api/dni/{numero}
  Consulta datos de personas naturales mediante su DNI (8 dígitos).
  Ejemplo: GET https://apiconsulta.sehuacho.com/api/dni/12345678

[GET] /api/ruc/{numero}
  Consulta información de empresas o contribuyentes (11 dígitos).
  Ejemplo: GET https://apiconsulta.sehuacho.com/api/ruc/20123456789


3. EJEMPLOS DE CÓDIGO
-----------------------

--- cURL ---
curl -X GET "https://apiconsulta.sehuacho.com/api/dni/12345678" \
     -H "x-api-key: tu_api_key_aqui"

--- Python ---
import requests

url = "https://apiconsulta.sehuacho.com/api/dni/12345678"
headers = {"x-api-key": "tu_api_key_aqui"}

response = requests.get(url, headers=headers)
print(response.json())

--- Go ---
package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
)

func main() {
    req, _ := http.NewRequest("GET", "https://apiconsulta.sehuacho.com/api/dni/12345678", nil)
    req.Header.Add("x-api-key", "tu_api_key_aqui")

    res, _ := http.DefaultClient.Do(req)
    defer res.Body.Close()
    body, _ := ioutil.ReadAll(res.Body)

    fmt.Println(string(body))
}

================================================
# API-KEY PARA ESTE PROYECTO
tc_a72ee220220a196d5eee19f9f1ba4e00
