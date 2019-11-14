package cmd

import (
	"strings"
	"fmt"
    "io/ioutil"
    "net/http"
    "time"
    jwt "github.com/dgrijalva/jwt-go"
)

//RunCommand -
func RunCommand(cmdName string) {
    validToken, err := GenerateJWT()
    if err != nil {
        fmt.Println("Failed to generate token")
    }
    client := &http.Client{}
    req, _ := http.NewRequest("GET", strings.Join([]string{Config.Serverurl, "run", cmdName}, "/"), nil)
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
    claims["client"] = "nsca-go client"
    claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

    tokenString, err := token.SignedString([]byte(Config.JwtKey))

    if err != nil {
        fmt.Errorf("Something Went Wrong: %s", err.Error())
        return "", err
    }

    return tokenString, nil
}
