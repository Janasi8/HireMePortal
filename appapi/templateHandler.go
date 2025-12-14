package signupapiv1

import (
	"html/template"
	"log"
	"net/http"
)

func templateHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html") // Path to your HTML file
	log.Printf("Template request: %s %s", r.Method, r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct{ Message string }{"Hello from Go!"}
	log.Print("Rendering template...", data, tmpl)
	tmpl.Execute(w, data)
}

// func templateLoginOrSignupHandler(w http.ResponseWriter, r *http.Request) {
// 	tmpl, err := template.ParseFiles("templates/loginorsignup.html") // Path to your HTML file
// 	log.Printf("Template request: %s %s", r.Method, r.URL.Path)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	data := struct{ Message string }{"Hello from Go!"}
// 	log.Print("Rendering template...", data, tmpl)
// 	tmpl.Execute(w, data)
// }

// func templateDashboardHandler(w http.ResponseWriter, r *http.Request) {
// 	tmpl, err := template.ParseFiles("templates/dashboard.html") // Path to your HTML file
// 	log.Printf("Template request: %s %s", r.Method, r.URL.Path)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	data := struct{ Message string }{"Hello from Go!"}
// 	log.Print("Rendering template...", data, tmpl)
// 	tmpl.Execute(w, data)
// }

func TestingHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/testing.html") // Path to your HTML file
	log.Printf("Template request: %s %s", r.Method, r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct{ Message string }{"Hello from Go!"}
	log.Print("Rendering template...", data, tmpl)
	tmpl.Execute(w, data)
}
