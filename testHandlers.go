package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type testResponse struct {
	PathVariable string `json:"pathVariable"`
	Method       string `json:"method"`
}

func getTestHandler(w http.ResponseWriter, _ *http.Request) {
	payload := `{"message":" /test , methods is GET"}`
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(payload))
	if err != nil {
		return
	}

}

func getTestHandlerWithPathVariable(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	response := testResponse{
		PathVariable: id,
		Method:       r.Method,
	}
	payload, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(payload)
}

func postTestHandler(w http.ResponseWriter, _ *http.Request) {
	payload := `{"message":" /test , methods is POST"}`
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(payload))
	if err != nil {
		return
	}

}
