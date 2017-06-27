FROM lmorg/level10fireball:latest

#####################################################################
# This packages up the code into a live deployable Docker container #
#####################################################################

# Pull this source
#RUN mkdir -p /go/src/github.com/lmorg
ADD . /go/src/github.com/lmorg/laurencemorgan.co.uk

# Compile the backend code
RUN go install github.com/lmorg/laurencemorgan.co.uk/level10fireball

# Compile frontend code
WORKDIR /go/src/github.com/lmorg/laurencemorgan.co.uk/site/
RUN sass scss/desktop.scss layout/desktop.css
RUN sass scss/mobile.scss  layout/mobile.css

# Minimize
#RUN mv layout/interactive.js layout/interactive-h.js && yui-compressor -o layout/interactive.js layout/interactive-h.js
#RUN mv layout/desktop.css    layout/desktop-h.css    && yui-compressor -o layout/desktop.css    layout/desktop-h.css
#RUN mv layout/mobile.css     layout/mobile-h.css     && yui-compressor -o layout/mobile.css     layout/mobile-h.css

# Make the site read-only aside the uploads path
#RUN chmod -R ugo-w /go/src/github.com/lmorg/laurencemorgan.co.uk

# Define web server
USER lvl10
ENTRYPOINT /go/bin/level10fireball --conf /go/src/github.com/lmorg/laurencemorgan.co.uk/site/conf/
EXPOSE 8080
