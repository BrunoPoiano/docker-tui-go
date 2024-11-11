
# Docker-TUI

Docker-tui is a simple terminal user interface to interact with docker. Writen in GO and CharmCLI 

![App Screenshot](./imgs/Docker-tui.png)



## Deploy

Clone the project

```bash
  git clone https://github.com/BrunoPoiano/docker-tui-go.git
```

cd to the directory

```bash
  cd docker-tui-go
```

Compile packages and dependencies

```bash
  go build .
```

Start the project

```bash
  go run .
```


## Docker Actions

Easiest way to start, stop and restart containers 

![Docker Actions](./imgs/menu.gif)


## Docker Logs

View the paginated logs 
![Docker Logs](./imgs/logs.gif)

## FAQ

#### Shell into containers

Shell functionality is there, it works, but it's still kinda broken. 
