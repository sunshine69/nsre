FROM golang:alpine as builder

# Example of build tiny docker image with nsre. However I prefer to build as
# static nsre linked with musl so I can run everywhere.

RUN apk update && \
    apk add --no-cache git ca-certificates tzdata gcc musl-dev && \
    update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

WORKDIR /work/

COPY . .

# Fetch dependencies.
# RUN go get -d -v

# Build the binary
#RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/hello .

RUN go build --ldflags '-extldflags "-static" -w -s'
#RUN go build --ldflags '-w -s'

############################
# STEP 2 build a small image
############################
FROM scratch

# Import from builder.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# Copy our static executable
COPY --from=builder /work/nsre /bin/nsre

# Use an unprivileged user.
USER appuser

# Run the hello binary.
ENTRYPOINT ["/bin/nsre"]
