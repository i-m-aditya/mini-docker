package main

import (

	// Uncomment this block to pass the first stage!

	"codecrafters-docker-go/app/client"
	"codecrafters-docker-go/app/utils"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"syscall"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage!
	//
	// fmt.Println(os.Args)
	image := os.Args[2]
	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	// create a directory
	chrootDir, err := os.MkdirTemp("", "chroot")

	if err != nil {
		fmt.Printf("error creating temp dir: %v", err)
		os.Exit(1)
	}

	tarDir, err := os.MkdirTemp("", "tarDir")
	if err != nil {
		fmt.Printf("error creating temp folder for tar files: %v", err)
		os.Exit(1)
	}

	defer cleanUp(chrootDir, tarDir)

	if err != nil {
		fmt.Printf("error creating temp dir: %v", err)
		os.Exit(1)
	}

	// fmt.Println(chrootDir, command)

	if err = copyExecutableToChrootDir(command, chrootDir); err != nil {
		fmt.Printf("error copying executable to chroot dir: %v", err)
		os.Exit(1)
	}
	paths := fetchDockerImageIntoDir(image, tarDir)

	err = ExtractTarsToDir(chrootDir, paths)
	if err != nil {
		fmt.Printf("error copying tars into chroot dir: %v", err)
		os.Exit(1)
	}

	if err = createDevNull(chrootDir); err != nil {
		fmt.Printf("error creating dev null: %v", err)
		os.Exit(1)
	}

	if err = syscall.Chroot(chrootDir); err != nil {
		fmt.Printf("chroot dir error: %v", err)
		os.Exit(1)
	}

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	syscall.Unshare(syscall.CLONE_NEWPID)
	//
	err = cmd.Run()

	exitErr, ok := err.(*exec.ExitError)

	if ok {
		if err = os.RemoveAll(chrootDir); err != nil {
			panic(err)
		}
		os.Exit(exitErr.ExitCode())
	} else if err != nil {
		if err = os.RemoveAll(chrootDir); err != nil {
			panic(err)
		}
		fmt.Printf("error running command: %v", err)
		os.Exit(1)
	}
}

func copyExecutableToChrootDir(executablePath string, chrootDir string) error {
	executablePathInChrootDir := path.Join(chrootDir, executablePath)
	// fmt.Println(path.Dir(executablePathInChrootDir))
	// fmt.Printf("%v , %v \n", executablePathInChrootDir, chrootDir)
	if err := os.MkdirAll(path.Dir(executablePathInChrootDir), 0750); err != nil {
		panic(err)
	}

	return copyFileToExecutable(executablePath, executablePathInChrootDir)

}
func cleanUp(chrootDir, tarDir string) {
	err := os.RemoveAll(tarDir)
	if err != nil {
		fmt.Printf("error removing tarDir: %v", err)
		os.Exit(1)
	}

	err = os.RemoveAll(chrootDir)
	if err != nil {
		fmt.Printf("error removing chrootDir: %v", err)
		os.Exit(1)
	}
}

func copyFileToExecutable(sourceFilePath string, destinationFilePath string) error {

	sourceFileMode, err := os.Stat(sourceFilePath)

	if err != nil {
		panic(err)
	}

	sourceFile, err := os.Open(sourceFilePath)

	if err != nil {
		panic(err)
	}

	defer sourceFile.Close()

	destinationFile, err := os.OpenFile(destinationFilePath, os.O_RDWR|os.O_CREATE, sourceFileMode.Mode())

	if err != nil {
		panic(err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return err

}

func createDevNull(chrootDir string) error {
	if err := os.MkdirAll(path.Join(chrootDir, "dev"), 0750); err != nil {
		return err
	}

	return ioutil.WriteFile(path.Join(chrootDir, "dev", "null"), []byte{}, 0644)
}

func fetchDockerImageIntoDir(image string, tarDir string) []string {
	repo, ref := utils.ParseImage(image)

	token := utils.GetAuthenticationToken(repo)

	dockerClient := client.GetNewDockerClient(
		repo,
		ref,
		token,
	)
	mainfest := dockerClient.PullManifest()

	_ = os.Mkdir(tarDir, 0755)
	// fmt.Println(tarDir)

	paths := dockerClient.PullLayers(mainfest, tarDir)

	return paths

}

func ExtractTarsToDir(chootDir string, paths []string) error {
	for _, path := range paths {
		cmd := exec.Command("tar", "xf", path, "-C", chootDir)
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}
