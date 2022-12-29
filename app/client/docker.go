package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

type DockerClient struct {
	repo  string
	ref   string
	token string
}

type manifestResponse struct {
	Name     string
	Tag      string
	FsLayers []struct {
		BlobSum string
	}
}

func GetNewDockerClient(repo string, ref string, token string) *DockerClient {
	return &DockerClient{
		repo:  repo,
		ref:   ref,
		token: token,
	}
}

func (dockerClient *DockerClient) PullManifest() manifestResponse {

	manifestUrl := fmt.Sprintf("https://registry.hub.docker.com/v2/%s/manifests/%s", dockerClient.repo, dockerClient.ref)

	response, err := dockerClient.authenticatedRequest(manifestUrl)

	if err != nil {
		fmt.Printf("Error getting manifest %v", err)
	}

	body, _ := ioutil.ReadAll(response.Body)

	var manifest manifestResponse
	err = json.Unmarshal(body, &manifest)
	if err != nil {
		fmt.Printf("Error %v", err)
	}

	// fmt.Println(manifest)
	return manifest
}

func (dockerClient *DockerClient) PullLayers(manifest manifestResponse, tarDir string) []string {

	var paths []string
	for _, layer := range manifest.FsLayers {
		layerUrl := fmt.Sprintf("https://registry.hub.docker.com/v2/%s/blobs/%s", dockerClient.repo, layer.BlobSum)
		fmt.Printf("Pulling layer Url %s\n", layerUrl)

		destPath := path.Join(tarDir, layer.BlobSum)
		err := dockerClient.PullLayer(destPath, layer.BlobSum)

		if err != nil {
			fmt.Printf("Error pulling layer %v", err)
		}
		paths = append(paths, destPath)
	}
	return paths
}

func (dockerClient *DockerClient) PullLayer(destPath, blobSum string) error {
	err := os.MkdirAll(path.Dir(destPath), 0750)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://registry.hub.docker.com/v2/%s/blobs/%s", dockerClient.repo, blobSum)
	resp, err := dockerClient.authenticatedRequest(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	io.Copy(f, resp.Body)
	return nil
}

func (dockerClient *DockerClient) authenticatedRequest(url string) (*http.Response, error) {

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Printf("Error creating request %v", err)
	}

	req.Header.Add(
		"Authorization", "Bearer "+dockerClient.token,
	)

	client := http.DefaultClient
	return client.Do(req)
}
