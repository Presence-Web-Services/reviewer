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

var config gmailer.Config
var status = http.StatusSeeOther

func init() {
  loadEnvVars()
  authenticate()
}

func CreateAndRun(port string) {
  http.HandleFunc("/", handler)
  http.ListenAndServe(":" + port, nil)
}

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
}

func authenticate() {
  err := config.Authenticate()
  if err != nil {
    log.Fatal("Error: Could not authenticate with GMail OAuth using credentials.")
  }
}

func sendEmail() {
  err := config.Send()
  if err != nil {
    status = http.StatusInternalServerError
    return
  }
}

func handler(response http.ResponseWriter, request *http.Request) {
  status = http.StatusSeeOther
  verifyPost(response, request.Method)
  if status != http.StatusSeeOther {
    response.WriteHeader(status)
    return
  }
  validate(request.Body)
  if status != http.StatusSeeOther {
    response.WriteHeader(status)
    return
  }
  sendEmail()
  if status != http.StatusSeeOther {
    response.WriteHeader(status)
    return
  }
  response.Header().Set("Location", "/email-sent")
  response.WriteHeader(status)
}

func verifyPost(response http.ResponseWriter, method string) {
  if method != "POST" {
    response.Header().Set("Allow", "POST")
    status = http.StatusMethodNotAllowed
  }
}

func validate(body io.ReadCloser) {
  var buffer bytes.Buffer
  var values url.Values
  if _, err := io.Copy(&buffer, body); err != nil {
    status = http.StatusInternalServerError
    return
  }
  bodyString := buffer.String()
  values, err := url.ParseQuery(bodyString)
  if err != nil {
    status = http.StatusInternalServerError
    return
  }
  checkBody(values)
}

func checkBody(values url.Values) {
  val, ok := values["hp"];
  if !ok || len(val) != 1 || val[0] != "" {
    status = http.StatusBadRequest
    return
  }
  val, ok = values["email"];
  if !ok || len(val) != 1 || !emailValid(val[0]) {
    status = http.StatusBadRequest
    return
  }
  config.ReplyTo = val[0]
  val, ok = values["message"];
  if !ok || len(val) != 1 || !messageValid(val[0]) {
    status = http.StatusBadRequest
    return
  }
  config.Body = val[0]
}

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

func messageValid(message string) bool {
  if len(message) == 0 || len(message) > 2000 {
    return false
  }
  return true
}
