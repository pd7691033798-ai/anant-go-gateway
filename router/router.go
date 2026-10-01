package router

import (
	"database/sql"
	"net/http"
)

// एक यूनिवर्सल इंटरफेस जिसे हर नया मॉड्यूल लागू करेगा
type ModuleRouter interface {
	RegisterRoutes(mux *http.ServeMux, db *sql.DB)
}

var registeredModules []ModuleRouter

// नए मॉड्यूल को खुद-ब-खुद रजिस्टर करने का फंक्शन
func RegisterModule(m ModuleRouter) {
	registeredModules = append(registeredModules, m)
}

// यह फंक्शन बिना main.go को बदले सभी मॉड्यूल्स के राउट्स को एक साथ जोड़ देगा
func InitAllRoutes(mux *http.ServeMux, db *sql.DB) {
	for _, mod := range registeredModules {
		mod.RegisterRoutes(mux, db)
	}
}
