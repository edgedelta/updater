FROM golang:1.19

WORKDIR /app

COPY . ./
RUN go mod download
RUN go build -o /agent-updater cmd/agent-updater/main.go

CMD [ "/agent-updater" ]