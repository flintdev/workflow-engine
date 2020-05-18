package healthCheck

import (
	"log"
	"net/http"
)

func homePage(w http.ResponseWriter, r *http.Request) {
}

func HealthCheck() {
	http.HandleFunc("/", homePage)
	log.Fatal(http.ListenAndServe(":9090", nil))
}
