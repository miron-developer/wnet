###################################################################################################
#                                                                                                 #
#                                   Miron-developer & MirasK                                      #
#                                   WNET - the social-network                                     #
#                                           (server)                                              #
###################################################################################################

FROM golang:1.13

COPY . .
WORKDIR /pkg
RUN go mod download; go build -o ./cmd/wnet cmd/main.go

LABEL description="This is the social-network project." \
    authors="Miron-developer, MirasK" \
    contacts="https://github.com/miron-developer, https://github.com/mirasK, wnet.soc.net@gmail.com" \
    site="https://wnet-sn.herokuapp.com"

CMD ["cmd/wnet"]

EXPOSE 4430