FROM golang:1.17.8-alpine3.15

RUN mkdir golangApplication

WORKDIR /golangApplication

COPY . .

RUN export GO111MODULE=on
RUN cd /golangApplication
RUN go build -o main.exe

EXPOSE 8080

CMD [ "/golangApplication/main.exe" ]

