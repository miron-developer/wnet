###################################################################################################
#                                                                                                 #
#                                   Miron-developer & MirasK                                      #
#                                   WNET - the social-network                                     #
#                                           (server)                                              #
###################################################################################################

FROM golang:1.13

COPY . .
WORKDIR /app
RUN go mod download; go build

LABEL description="This is the social-network project." \
    authors="Miron-developer, MirasK" \
    contacts="https://github.com/miron-developer, https://github.com/mirasK, wnet.soc.net@gmail.com" \
    site="https://wnet.netlify.app"

CMD ["./wnet"]

EXPOSE 8080
EXPOSE 4430