# mini-docker

> An lightweight Go implementation of Docker's `run` command.

mini docker is an implementation of a very basic Docker runner (which allows you to execute arbitrary commands on any Docker image)

## Structure

The codebase is structured as follows:

```
mini-docker/
├─ app/
│ ├─ main.go: What gets called when you run the project
│ ├─ client/
│ │ ├─ docker.go: For authenticating and interacting with registry
│ ├─ utils
│ │ ├─ dockerUtil.go: Utility file for especially parsing docker responses
├─ README.md
```


## Usage

Since running the program requires root (in order to properly isolate the processes), we're running our Docker implementation _inside_ of a Docker container. To make things easier, you can declare the following alias

```bash
alias mini_docker='docker build -t mini_docker . && docker run --cap-add="SYS_ADMIN" mini_docker'
```

Then, run commands as you would with docker: `mini_docker run alpine:latest echo hi`
