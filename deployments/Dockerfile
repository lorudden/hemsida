FROM --platform=$BUILDPLATFORM docker.io/golang:1.26.3 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

RUN go install github.com/a-h/templ/cmd/templ@latest

ARG BUILDARCH

RUN echo $BUILDARCH

RUN curl --fail -sL "https://github.com/tailwindlabs/tailwindcss/releases/download/v4.3.0/tailwindcss-linux-$BUILDARCH" -o tailwindcss
RUN chmod +x tailwindcss

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN templ generate
RUN ./tailwindcss -i ./internal/platform/web/css/input.css -o ./assets/css/loran.css

WORKDIR /app/cmd/hemsida

ARG TARGETARCH
ARG TARGETOS

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build


FROM scratch

WORKDIR /opt/lorudden

COPY --from=builder --chown=65532 /app/cmd/hemsida/hemsida /opt/lorudden
COPY --from=builder --chown=65532 /app/assets /opt/lorudden/assets
COPY --from=builder --chown=65532 /app/data /opt/lorudden/data

USER 65532:65532

EXPOSE 8080
ENTRYPOINT ["/opt/lorudden/hemsida"]
