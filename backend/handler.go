package main

import (
	"encoding/json"
	"net/http"
)

type RequestBody struct {
	Message string `json:"message"`
}

type ResponseBody struct {
	Message string `json:"message"`
}

func messageHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RequestBody
	err := json.NewDecoder(request.Body).Decode(&req)
	if err != nil {
		http.Error(writer, "Invalid JSON", http.StatusBadRequest)
		return
	}
	resp := ResponseBody{Message: "Got, " + req.Message}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(resp)
}
