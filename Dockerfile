FROM golang

# Pull this repo
RUN mkdir -p /go/src/github.com/lmorg
RUN git clone https://github.com/lmorg/laurencemorgan.co.uk /go/src/github.com/lmorg/laurencemorgan.co.uk

# Get Go dependancies
RUN go get -t -u github.com/go-sql-driver/mysql
RUN go get -t -u github.com/nfnt/resize
RUN go get -t -u golang.org/x/crypto/scrypt
RUN go get -t -u golang.org/x/net/websocket

# Compile the backend code
RUN go install github.com/lmorg/laurencemorgan.co.uk/level10fireball

# Start webserver
ENTRYPOINT /go/bin/level10fireball --conf /go/src/github.com/lmorg/laurencemorgan.co.uk/site/conf

EXPOSE 8080
