# Starting the app
First, cd to the directory to clone the repo into

```cd /directory/for/app```

Clone the repo with

```git clone https://github.com/StingrayFG/buffer_test```

And cd into it

```cd buffer_test```

## Development environment
In order to start the app in the development environment, create a .env file or copy the .env.example file with

```cp .env.example .env```

After that download dependencies

```go mod download```

And run it in with

```go run main.go```

## Docker container
In order to start the app as a docker container, build the container with

```sudo docker build --tag buffer_test .```

Then run it with docker run

```sudo docker run --name test -d -p 5500 buffer_test```

Or using docker compose

```sudo docker compose up```





