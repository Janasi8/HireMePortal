package signupapiv1

import (
	dbpg "achievesomethingbro/appdb"
	localmodel "achievesomethingbro/appmodel"
	"html/template"
	"log"
	"net/http"
)

// adminHandler handles the login form display and submission.
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("admin")
	log.Print(r.URL.Path)
	if r.Method == http.MethodGet {
		log.Print("admin1")
		tmpl, _ := template.ParseFiles("templates/admin.html")
		tmpl.Execute(w, nil)
	} else if r.Method == http.MethodPost {
		log.Print("admin2")
		r.ParseForm()
		// username := r.FormValue("username")
		// password := r.FormValue("password")

		// var storedPassword string
		// db, err := dbpg.ConnectDB()
		// if err != nil {
		// 	http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		// 	return
		// }
		// // Check if the admin exists in the database.
		// err = db.QueryRow("SELECT password FROM admins WHERE username = ?", username).Scan(&storedPassword)

		// if err != nil || password != storedPassword {
		// 	// Authentication failed.
		// 	fmt.Fprintf(w, "Invalid credentials. <a href='/admin'>Try again</a>")
		// 	return
		// }

		// Authentication successful, show the dashboard.
		showDashboard(w, r)
	}
}

// showDashboard fetches all users and displays them.
func showDashboard(w http.ResponseWriter, r *http.Request) {
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT id, user_name, password FROM users")
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []localmodel.User
	for rows.Next() {
		var u localmodel.User
		if err := rows.Scan(&u.ID, &u.UserName, &u.Password); err != nil {
			http.Error(w, "Failed to scan user row", http.StatusInternalServerError)
			return
		}
		log.Printf("User: %+v\n", u)
		users = append(users, u)
	}

	data := localmodel.PageData{
		Users: users,
	}
	log.Print("admindashboar")
	tmpl, _ := template.ParseFiles("templates/admin.html")
	tmpl.Execute(w, data)
}
