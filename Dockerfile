FROM golang:latest 

ADD main.go /main.go
ADD .env /.env
CMD ["go", "run", "/main.go"]
