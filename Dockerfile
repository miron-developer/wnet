###################################################################################################
#                                                                                                 #
#                                   Miron-developer & MirasK                                      #
#                                 AniFor - the real-time-forum                                    #
#                                                                                                 #
###################################################################################################

FROM golang:1.13

COPY . .
WORKDIR /app
RUN go mod download; go build

LABEL description="This is the real-time-forum project." \
    authors="Miron-developer, MirasK" \
    contacts="https://github.com/miron-developer, https://github.com/mirasK" \
    site="https://anifor.herokuapp.com"

CMD ["./anifor"]

EXPOSE 8080
EXPOSE 4430