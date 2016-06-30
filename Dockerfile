FROM golang:1.6

ADD . /go/src/github.com/ren-/score-whisperer

RUN go get github.com/bwmarrin/discordgo
WORKDIR /go/src/github.com/bwmarrin/discordgo
RUN git checkout develop


WORKDIR /go/src/github.com/ren-/score-whisperer
RUN go get ./...

WORKDIR /go/src/github.com/ren-/score-whisperer/cmd/whisperer
RUN go install

ENTRYPOINT ["/go/bin/whisperer", "--discordowneruserid", "137214686973132800",  "--discordapplicationclientid", "189474870923362305"]


