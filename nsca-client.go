package main

import (
	"flag"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"
    jwt "github.com/dgrijalva/jwt-go"
)

var mySigningKey = []byte("kGay08Hf5KvSIhYREkiq2FJYNstQsrTK")

func SendRequest(cmdName string) {
    validToken, err := GenerateJWT()
    if err != nil {
        fmt.Println("Failed to generate token")
    }

    client := &http.Client{}
    req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/run/%s", cmdName), nil)
    req.Header.Set("Token", validToken)
    res, err := client.Do(req)
    if err != nil {
        fmt.Printf("Error: %v", err)
    }

    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Printf("%s", string(body))
}

func GenerateJWT() (string, error) {
    token := jwt.New(jwt.SigningMethodHS256)

    claims := token.Claims.(jwt.MapClaims)

    claims["authorized"] = true
    claims["client"] = "Elliot Forbes"
    claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

    tokenString, err := token.SignedString(mySigningKey)

    if err != nil {
        fmt.Errorf("Something Went Wrong: %s", err.Error())
        return "", err
    }

    return tokenString, nil
}

func main() {
	cmdName := flag.String("cmd", "", "Command name")
	flag.Parse()
	SendRequest(*cmdName)
}