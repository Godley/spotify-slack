FROM golang:latest 

ADD main.go /main.go
ADD .env /.env
RUN source .env
CMD ["go", "run", "/main.go"]
