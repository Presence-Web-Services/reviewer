/*
Inquirer sets up a server that listens for POST data (on specified port) and sends email using gmailer if POST data is valid.
*/
package inquirer

import (
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/presence-web-services/gmailer/v2"
	"github.com/joho/godotenv"
)

// gmailer config for sending email
var config gmailer.Config

// default important values
var status = http.StatusOK
var errorMessage = ""
var hp = ""
var site = ""

// init loads environment variables and authenticates the gmailer config
func init() {
	loadEnvVars()
	authenticate()
}

// CreateAndRun is exported to allow for creation of an inquirer
func CreateAndRun(port string) {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+port, nil)
}

// loadEnvVars loads environment variables from a .env file
func loadEnvVars() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error: Could not load environment variables from .env file.")
	}
	config.ClientID = os.Getenv("CLIENT_ID")
	config.ClientSecret = os.Getenv("CLIENT_SECRET")
	config.AccessToken = os.Getenv("ACCESS_TOKEN")
	config.RefreshToken = os.Getenv("REFRESH_TOKEN")
	config.EmailTo = os.Getenv("EMAIL_TO")
	config.EmailFrom = os.Getenv("EMAIL_FROM")
	config.Subject = os.Getenv("SUBJECT")
	site = "https://" + os.Getenv("SITE")
}

// authenticate authenticates a gmailer config
func authenticate() {
	err := config.Authenticate()
	if err != nil {
		log.Fatal("Error: Could not authenticate with GMail OAuth using credentials.")
	}
}

// sendEmail sends an email given a gmailer config
func sendEmail() {
	err := config.Send()
	if err != nil {
		status = http.StatusInternalServerError
		errorMessage = "Error: Internal server error."
		return
	}
}

// defaultValues sets the status, errorMessage, ReplyTo, Body all to default values
func defaultValues() {
	status = http.StatusOK
	errorMessage = ""
	hp = ""
	config.ReplyTo = ""
	config.Body = ""
}

// handler verifies a POST is sent, and then validates the POST data, and sends an email if valid
func handler(response http.ResponseWriter, request *http.Request) {
	defaultValues()
	response.Header().Set("Access-Control-Allow-Origin", site)
	checkOrigin(request)
	if status != http.StatusOK {
		http.Error(response, errorMessage, status)
		return
	}
	verifyPost(response, request.Method)
	if status != http.StatusOK {
		http.Error(response, errorMessage, status)
		return
	}
	getFormData(request)
	checkEmail()
	if status != http.StatusOK {
		http.Error(response, errorMessage, status)
		return
	}
	checkMessage()
	if status != http.StatusOK {
		http.Error(response, errorMessage, status)
		return
	}
	checkHP()
	if status != http.StatusOK {
		http.Error(response, errorMessage, status)
		return
	}
	sendEmail()
	if status != http.StatusOK {
		http.Error(response, errorMessage, status)
		return
	}
	response.Write([]byte("Email sent successfully!"))
}

// checkOrigin ensures origin is from proper website
func checkOrigin(request *http.Request) {
	origin := request.Header.Get("Origin")
	if origin != site {
		status = http.StatusForbidden
		errorMessage = "Error: Only certain sites are allowed to use this endpoint."
		return
	}
}

// verifyPost ensures that a POST is sent
func verifyPost(response http.ResponseWriter, method string) {
	if method != "POST" {
		response.Header().Set("Allow", "POST")
		status = http.StatusMethodNotAllowed
		errorMessage = "Error: Method " + method + " not allowed. Only POST allowed."
	}
}

// getFormData populates config struct and hp variable with POSTed data from form submission
func getFormData(request *http.Request) {
	config.ReplyTo = request.PostFormValue("email")
	config.Body = request.PostFormValue("message")
	hp = request.PostFormValue("hp")
}

// checkEmail verifies email submitted is valid
func checkEmail() {
	if len(config.ReplyTo) < 5 || len(config.ReplyTo) > 50 {
		status = http.StatusBadRequest
		errorMessage = "Error: Email is too short or too long."
		return
	}
	var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !emailRegex.MatchString(config.ReplyTo) {
		status = http.StatusBadRequest
		errorMessage = "Error: Email is not a valid format."
		return
	}
	domain := strings.Split(config.ReplyTo, "@")[1]
	mx, err := net.LookupMX(domain)
	if err != nil || len(mx) == 0 {
		status = http.StatusBadRequest
		errorMessage = "Error: Domain given is not a valid email domain."
		return
	}
}

// checkMessage verifies message submitted is valid
func checkMessage() {
	if len(config.Body) == 0 || len(config.Body) > 2000 {
		status = http.StatusBadRequest
		errorMessage = "Error: Message is too long or empty."
		return
	}
}

// checkHP ensures honeypot field is not populated
func checkHP() {
	if hp != "" {
		status = http.StatusBadRequest
		errorMessage = "Error: Please, no robots!"
	}
}
