package signupapiv1

// THIS FILE IS INTENTIONALLY DISABLED.
//
// Reason:
// We are using GIN for routing and HTML rendering in main.go.
// Old net/http + template.ParseFiles handlers caused
// dashboard and API routes to be overridden.
//
// DO NOT add any code here unless you fully remove Gin.
// All templates are now served using:
//
//   router.LoadHTMLGlob("templates/*")
//   router.GET("/dashboard", ...)
//   router.GET("/", ...)
