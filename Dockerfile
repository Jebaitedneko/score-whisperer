FROM golang:1.6

ADD . /go/src/github.com/ren-/score-whisperer

RUN go get github.com/bwmarrin/discordgo
WORKDIR /go/src/github.com/bwmarrin/discordgo
RUN git checkout develop


WORKDIR /go/src/github.com/ren-/score-whisperer
RUN go get ./...

WORKDIR /go/src/github.com/ren-/score-whisperer/cmd/whisperer
RUN go install

ENTRYPOINT ["/go/bin/whisperer", "--discordowneruserid", "110076057167618048",  "--discordapplicationclientid", "254494841201754112"]


