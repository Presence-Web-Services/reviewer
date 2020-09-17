/*
Inquirer sets up a server that listens for POST data (on specified port) and sends email using gmailer if POST data is valid.
*/
package inquirer

import (
  "io"
  "bytes"
  "regexp"
  "strings"
  "net"
  "os"
  "log"
  "net/http"
  "net/url"

  "github.com/joho/godotenv"
  "github.com/Presence-Web-Services/gmailer/v2"
)

// gmailer config for sending email
var config gmailer.Config
// default status and error message
var status = http.StatusOK
var errorMessage = ""

// init loads environment variables and authenticates the gmailer config
func init() {
  loadEnvVars()
  authenticate()
}

// CreateAndRun is exported to allow for creation of an inquirer
func CreateAndRun(port string) {
  http.HandleFunc("/", handler)
  http.ListenAndServe(":" + port, nil)
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
  if os.Getenv("HP") == "true" {
    config.HP = true
  } else {
    config.HP = false
  }
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
  config.ReplyTo = ""
  config.Body = ""
}

// handler verifies a POST is sent, and then validates the POST data, and sends an email if valid
func handler(response http.ResponseWriter, request *http.Request) {
  defaultValues()
  verifyPost(response, request.Method)
  if status != http.StatusOK {
    http.Error(response, errorMessage, status)
    return
  }
  validate(request.Body)
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

// verifyPost ensures that a POST is sent
func verifyPost(response http.ResponseWriter, method string) {
  if method != "POST" {
    response.Header().Set("Allow", "POST")
    status = http.StatusMethodNotAllowed
    errorMessage = "Error: Method " + method + " not allowed. Only POST allowed."
  }
}

// validate validates the POST data
func validate(body io.ReadCloser) {
  var buffer bytes.Buffer
  var values url.Values
  if _, err := io.Copy(&buffer, body); err != nil {
    status = http.StatusInternalServerError
    errorMessage = "Error: Internal Server error."
    return
  }
  bodyString := buffer.String()
  values, err := url.ParseQuery(bodyString)
  if err != nil {
    status = http.StatusInternalServerError
    errorMessage = "Error: Internal Server error."
    return
  }
  checkBody(values)
}

// checkBody ensures the POST data body is well-formed
func checkBody(values url.Values) {
  expectedLength := 2
  if config.HP {
    expectedLength++;
  }
  if len(values) != expectedLength {
    status = http.StatusBadRequest
    errorMessage = "Error: Bad request."
    return
  }
  val, ok := values["hp"];
  if config.HP && (!ok || len(val) != 1 || val[0] != "") {
    status = http.StatusBadRequest
    errorMessage = "Error: Bad request."
    return
  }
  val, ok = values["email"];
  if !ok || len(val) != 1 || !emailValid(val[0]) {
    status = http.StatusBadRequest
    errorMessage = "Error: Bad request. Email may be too long or not valid."
    return
  }
  config.ReplyTo = val[0]
  val, ok = values["message"];
  if !ok || len(val) != 1 || !messageValid(val[0]) {
    status = http.StatusBadRequest
    errorMessage = "Error: Bad request. Message may be too long."
    return
  }
  config.Body = val[0]
}

// emailValid checks for email length, regex, and if valid MX domain
func emailValid(email string) bool {
  if len(email) < 5 || len(email) > 50 {
    return false
  }
  var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
  if !emailRegex.MatchString(email) {
    return false
  }
  domain := strings.Split(email, "@")[1]
  mx, err := net.LookupMX(domain)
  if err != nil || len(mx) == 0 {
    return false
  }
  return true
}

// messageValid checks message length
func messageValid(message string) bool {
  if len(message) == 0 || len(message) > 2000 {
    return false
  }
  return true
}
