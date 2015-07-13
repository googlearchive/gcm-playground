FROM golang:1.4.2-wheezy


# -----------------
# Install dependencies
# -----------------

RUN apt-get update && apt-get install -y build-essential


# -----------------
# Install Go dependencies
# -----------------

RUN go get github.com/gorilla/mux

# -----------------
# Copy files over
# -----------------

ADD . /go/src/gcm-app-server
WORKDIR /go/src/gcm-app-server


# -----------------
# Set the server
# -----------------

EXPOSE 8080
