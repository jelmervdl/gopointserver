FROM golang:1.9
WORKDIR /go/src/github.com/jelmervdl/gopointserver
COPY . .
RUN go-wrapper download
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' .

FROM scratch
COPY --from=0 /go/src/github.com/jelmervdl/gopointserver/gopointserver .
EXPOSE 8000
VOLUME "/data"
ENTRYPOINT ["./gopointserver", "/data/*.geojson"]