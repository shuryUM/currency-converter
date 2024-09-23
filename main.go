package main

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
)

type Currency struct {
	Code string  `json:"code"`
	Rate float64 `json:"rate"` // Taxa de conversão para USD
}

var (
	currencies = []Currency{}
	mu         sync.Mutex
)

func main() {
	router := httprouter.New()
	router.GET("/currencies", getCurrencies)
	router.POST("/currencies", addCurrency)
	router.PUT("/currencies", updateCurrency)
	router.DELETE("/currencies", deleteCurrency)
	router.POST("/convert", convertCurrency)

	http.ListenAndServe(":8080", router)
}

func getCurrencies(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(currencies)
}

func addCurrency(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	mu.Lock()
	defer mu.Unlock()
	var newCurrency Currency
	json.NewDecoder(r.Body).Decode(&newCurrency)
	currencies = append(currencies, newCurrency)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newCurrency)
}

func updateCurrency(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	mu.Lock()
	defer mu.Unlock()
	var updatedCurrency Currency
	json.NewDecoder(r.Body).Decode(&updatedCurrency)
	for i, currency := range currencies {
		if currency.Code == updatedCurrency.Code {
			currencies[i] = updatedCurrency
			json.NewEncoder(w).Encode(updatedCurrency)
			return
		}
	}
	http.NotFound(w, r)
}

func deleteCurrency(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	mu.Lock()
	defer mu.Unlock()
	var currencyToDelete Currency
	json.NewDecoder(r.Body).Decode(&currencyToDelete)
	for i, currency := range currencies {
		if currency.Code == currencyToDelete.Code {
			currencies = append(currencies[:i], currencies[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.NotFound(w, r)
}

func convertCurrency(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	mu.Lock()
	defer mu.Unlock()
	var request struct {
		From   string  `json:"from"`
		To     string  `json:"to"`
		Amount float64 `json:"amount"`
	}
	json.NewDecoder(r.Body).Decode(&request)

	fromRate := getRate(request.From)
	toRate := getRate(request.To)

	if fromRate == 0 || toRate == 0 {
		http.Error(w, "Invalid currency code", http.StatusBadRequest)
		return
	}

	convertedAmount := (request.Amount / fromRate) * toRate
	json.NewEncoder(w).Encode(map[string]float64{"convertedAmount": convertedAmount})
}

func getRate(code string) float64 {
	for _, currency := range currencies {
		if currency.Code == code {
			return currency.Rate
		}
	}
	return 0
}
