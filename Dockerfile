FROM golang

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o /server ./cmd/server/main.go

CMD ["/server"]