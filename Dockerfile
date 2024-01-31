FROM golang:1.21.6-alpine3.19

COPY . /code

RUN cd /code && go mod tidy && go install ./cmd/gsproxy

ENTRYPOINT ["gsproxy"]
CMD ["-h"]