############################
# STEP 1 build executable binary
############################

FROM golang:alpine as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

COPY frontend/ $GOPATH/src/abv/frontend/
WORKDIR $GOPATH/src/abv/frontend/

# Copy the static content
RUN mkdir -p /srv/http/static
COPY frontend/front.html /srv/http
COPY frontend/static /srv/http/static

# Fetch dependencies.
RUN go get -d -v

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/frontend github.com/bhutch29/abv/frontend/

############################
# STEP 2 build a small image
############################

FROM scratch

# Import from builder.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/group /etc/shadow /etc/

# Copy our static executable
COPY --from=builder /go/bin/frontend /go/bin/frontend

# Use an unprivileged user.
USER appuser

# Copy config file
ADD --chown=appuser config.toml /home/appuser/.abv/

COPY --from=builder /srv/http /srv/http

EXPOSE 8080

ENTRYPOINT ["/go/bin/frontend"]
