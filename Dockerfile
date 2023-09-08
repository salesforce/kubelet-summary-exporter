FROM golang:1.21 as build

WORKDIR /go/src/app

COPY go.* .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 make install
  
FROM gcr.io/distroless/static-debian11:nonroot
COPY --from=build /go/bin/kubelet-summary-exporter /
CMD ["/kubelet-summary-exporter"]
