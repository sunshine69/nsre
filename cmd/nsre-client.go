package cmd

import (
	"crypto/tls"
	"strconv"
	"os"
	"strings"
	"fmt"
    "io/ioutil"
    "net/http"
    "time"
    jwt "github.com/dgrijalva/jwt-go"
)

//RunCommand -
func RunCommand(cmdName ...string) (string) {
    validToken, err := GenerateJWT()
    if err != nil {
        fmt.Printf("ERROR - Failed to generate token - %v\n", err)
    }

    client := &http.Client{}


    var serverURL string
    if len(cmdName) == 3 {
        serverURL = cmdName[2]
    } else {
        serverURL = Config.Serverurl
    }

    if strings.HasPrefix(serverURL, "https://"){
        http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: Config.IgnoreCertificateCheck}
    }

    req, _ := http.NewRequest("GET", strings.Join([]string{serverURL, "run", cmdName[0]}, "/"), nil)
    req.Header.Set("Token", validToken)

    res, err := client.Do(req)
    if err != nil {
        fmt.Printf("ERROR - %v", err)
    }
    defer res.Body.Close()

    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        fmt.Printf("ERROR - %v", err)
    }

    o := string(body)
    if len(cmdName) == 1 {
        fmt.Printf("%s", o)
    }
    return o
}

//RunNagiosCheckCommand - This will return based on nagios plugin specs
func RunNagiosCheckCommand(cmdName string) {
    o := RunCommand(cmdName, "quiet")
    _splited := strings.SplitN(o, "\n", 2)
    fmt.Print(_splited[1])
    c, e := strconv.Atoi(_splited[0])
    if e != nil {
        fmt.Printf("ERROR can not parse exit code from remote - %v", e)
        os.Exit(2)
    }
    os.Exit(c)
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
