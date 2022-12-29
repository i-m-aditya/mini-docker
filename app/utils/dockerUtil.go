package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type DockerAuthResponse struct {
	Token string
}

func ParseImage(image string) (string, string) {
	parts := strings.Split(image, ":")
	repo := "library/" + parts[0]
	ref := "latest"
	if len(parts) == 2 {
		ref = parts[1]
	}
	return repo, ref
}

func GetAuthenticationToken(repo string) string {

	authUrl := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", repo)
	response, err := http.Get(authUrl)

	if err != nil {
		fmt.Printf("Error getting authentication token %v\n", err)
	}

	// fmt.Println(response.Status)
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		fmt.Printf("Error reading response body %v\n", err)
	}
	// fmt.Println(body)

	var authResp DockerAuthResponse

	err = json.Unmarshal(body, &authResp)

	if err != nil {

		fmt.Printf("Error %v", err)
	}

	// fmt.Println(authResp.Token)
	//
	return authResp.Token
}
