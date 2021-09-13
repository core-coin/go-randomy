FROM golang:alpine
RUN apk add g++ gcc
WORKDIR go-randomy
COPY . go-randomy/ 
RUN cd go-randomy && go test -v
