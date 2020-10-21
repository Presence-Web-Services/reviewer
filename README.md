# inquirer

Setup .env file:
```
CLIENT_ID=your_client_id
CLIENT_SECRET=your_client_secret
ACCESS_TOKEN=your_access_token
REFRESH_TOKEN=your_refresh_token

EMAIL_TO=to_email
EMAIL_FROM=from_email
SUBJECT="Your Subject Here"
SITE="yoursite.com"
```

How to run:
```
go mod download
go build -o ./inquirer-server
./inquirer-server
```

Running in Docker container:
```
docker build -t inquirer .
docker run -p 80:80 -d inquirer
```
