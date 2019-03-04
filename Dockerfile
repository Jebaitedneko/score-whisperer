FROM golang:1.12

ADD . /go/src/github.com/ren-/score-whisperer

RUN go get github.com/jonas747/discordgo
WORKDIR /go/src/github.com/jonas747/discordgo
RUN git checkout master


WORKDIR /go/src/github.com/ren-/score-whisperer
RUN go get ./...

WORKDIR /go/src/github.com/ren-/score-whisperer/cmd/whisperer
RUN go install

CMD ["/go/bin/whisperer", "--discordowneruserid", "110076057167618048",  "--discordapplicationclientid", "254494841201754112"]


