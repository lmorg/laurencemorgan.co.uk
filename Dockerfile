FROM golang

## Update host
#RUN apt-get update
#RUN apt-get -y upgrade

# Pull this source
RUN mkdir -p /go/src/github.com/lmorg
ADD . /go/src/github.com/lmorg/laurencemorgan.co.uk

# Get Go dependancies
RUN go get -t -u github.com/go-sql-driver/mysql
RUN go get -t -u github.com/nfnt/resize
RUN go get -t -u golang.org/x/crypto/scrypt
RUN go get -t -u golang.org/x/net/websocket
RUN go get -t -u github.com/kardianos/osext

# Compile the backend code
RUN go install github.com/lmorg/laurencemorgan.co.uk/level10fireball

## Compile frontend code
#RUN apt-get install -y ruby
#RUN gem install sass

# Uploads directory
RUN mkdir /uploads

## Copy the database.json file
#RUN mkdir /secrets
#RUN ln -s /secrets/database.json /go/src/github.com/lmorg/laurencemorgan.co.uk/site/conf/database.json

# Make the site read-only aside the uploads path
#RUN chmod -R ugo-w /go/src/github.com/lmorg/laurencemorgan.co.uk

# Start webserver
ENTRYPOINT /go/bin/level10fireball --conf /go/src/github.com/lmorg/laurencemorgan.co.uk/site/conf/

EXPOSE 8080
