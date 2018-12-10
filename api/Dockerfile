############################
# STEP 1 build executable binary
############################

FROM golang:alpine as builder

# Install git + SSL ca certificates + gcc.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
# gcc and musl-dev are required for building sqlite3 package
RUN apk update && apk add --no-cache git ca-certificates gcc musl-dev

# Create appuser
RUN adduser -D -g '' appuser

COPY api/ $GOPATH/src/abv/api/
WORKDIR $GOPATH/src/abv/api/

# Fetch dependencies.
RUN go get -d -v

# Build the binary
RUN go build -o /go/bin/api github.com/bhutch29/abv/api/

############################
# STEP 2 build a small image
############################

FROM alpine

# Import from builder.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/group /etc/shadow /etc/

# Copy our static executable
COPY --from=builder /go/bin/api /go/bin/api

# Use an unprivileged user.
USER appuser

# Copy config file
ADD --chown=appuser config.toml /home/appuser/.abv/

EXPOSE 8081

# Run the hello binary.
ENTRYPOINT ["/go/bin/api"]
