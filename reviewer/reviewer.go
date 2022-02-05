/*
Reviewer sets up a server that listens for POST data (on specified port) and sends email using gmailer if POST data is valid.
*/
package reviewer

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
var name = ""
var stars = ""
var review = ""

// init loads environment variables and authenticates the gmailer config
func init() {
	loadEnvVars()
	authenticate()
}

// CreateAndRun is exported to allow for creation of a reviewer
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
  name = ""
  stars = ""
  review = ""
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
  checkName()
  if status != http.StatusOK {
		http.Error(response, errorMessage, status)
		return
	}
	checkEmail()
	if status != http.StatusOK {
		http.Error(response, errorMessage, status)
		return
	}
  checkRating()
  if status != http.StatusOK {
		http.Error(response, errorMessage, status)
		return
	}
	checkReview()
	if status != http.StatusOK {
		http.Error(response, errorMessage, status)
		return
	}
	checkHP()
	if status != http.StatusOK {
		http.Error(response, errorMessage, status)
		return
	}
  createBody()
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
	name = request.PostFormValue("name")
  stars = request.PostFormValue("stars")
  review = request.PostFormValue("review")
	hp = request.PostFormValue("hp")
}

// checkName verifies name is valid length
func checkName() {
  if len(name) == 0 || len(name) > 100 {
    status = http.StatusBadRequest
		errorMessage = "Error: Name is blank or too long."
		return
  }
}

func checkStars() {
  stars_int, err := strconv.Atoi(stars)
  if err || stars_int < 1 || stars_int > 5 {
    status = http.StatusBadRequest
		errorMessage = "Error: Star rating must be between 1-5."
		return
  }
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

// checkReview verifies message submitted is valid
func checkReview() {
	if len(review) == 0 || len(review) > 2000 {
		status = http.StatusBadRequest
		errorMessage = "Error: Review is too long or empty."
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

func createBody() {
  config.Body = fmt.Sprintf("Name: %s\nStars: %s\nReview: %s", name, stars, review)
}
