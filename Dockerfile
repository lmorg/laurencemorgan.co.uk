FROM golang

## Update host
RUN apt-get update
RUN apt-get install -y ruby yui-compressor
RUN gem install sass
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

# Compile frontend code
WORKDIR /go/src/github.com/lmorg/laurencemorgan.co.uk/site/
RUN sass scss/desktop.scss layout/desktop-h.css
RUN sass scss/mobile.scss  layout/mobile-h.css

# Minimize
RUN mv layout/interactive.js layout/interactive-h.js
RUN yui-compressor -o layout/iteractive.js  layout/interactive-h.js
RUN yui-compressor -o layout/desktop.css layout/desktop-h.css
RUN yui-compressor -o layout/mobile.css  layout/mobile-h.css

# Uploads directory
RUN mkdir /uploads

# Make the site read-only aside the uploads path
#RUN chmod -R ugo-w /go/src/github.com/lmorg/laurencemorgan.co.uk

RUN groupadd -r lvl10 && useradd --no-log-init -r -g lvl10 lvl10
USER lvl10

# Start webserver
ENTRYPOINT /go/bin/level10fireball --conf /go/src/github.com/lmorg/laurencemorgan.co.uk/site/conf/

EXPOSE 8080
