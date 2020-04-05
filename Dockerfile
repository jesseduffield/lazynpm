# run with:
# docker build -t lazynpm .
# docker run -it lazynpm:latest /bin/sh -l

FROM golang:1.14-alpine3.11
WORKDIR /go/src/github.com/jesseduffield/lazynpm/
COPY ./ .
RUN CGO_ENABLED=0 GOOS=linux go build

FROM alpine:3.11
RUN apk add -U git xdg-utils
WORKDIR /go/src/github.com/jesseduffield/lazynpm/
COPY --from=0 /go/src/github.com/jesseduffield/lazynpm /go/src/github.com/jesseduffield/lazynpm
COPY --from=0 /go/src/github.com/jesseduffield/lazynpm/lazynpm /bin/
RUN echo "alias gg=lazynpm" >> ~/.profile
