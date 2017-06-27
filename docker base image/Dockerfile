FROM golang

#####################################################################
# This packages up the code into a live deployable Docker container #
#####################################################################

# Update host
RUN apt-get update
# (update required for Ruby)
RUN apt-get install -y ruby yui-compressor
RUN gem install sass

# Get Go dependancies
RUN go get -t -u github.com/go-sql-driver/mysql
RUN go get -t -u github.com/nfnt/resize
RUN go get -t -u golang.org/x/crypto/scrypt
RUN go get -t -u golang.org/x/net/websocket
RUN go get -t -u github.com/kardianos/osext

# Make source directories
RUN mkdir /uploads

RUN groupadd -r lvl10 && useradd --no-log-init -r -g lvl10 lvl10