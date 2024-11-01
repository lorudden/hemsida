FROM --platform=$BUILDPLATFORM docker.io/golang:1.23 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

RUN go install github.com/a-h/templ/cmd/templ@v0.2.771

ARG BUILDARCH

RUN echo $BUILDARCH

RUN curl --fail -sL "https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.13/tailwindcss-linux-$BUILDARCH" -o tailwindcss
RUN chmod +x tailwindcss

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN templ generate
RUN ./tailwindcss -c ./internal/pkg/presentation/web/css/tailwind.config.js -i ./internal/pkg/presentation/web/css/input.css -o ./assets/css/loran.css

WORKDIR /app/cmd/hemsida

ARG TARGETARCH
ARG TARGETOS

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build


FROM registry.access.redhat.com/ubi9/ubi-minimal
WORKDIR /opt/lorudden

COPY --from=builder --chown=1001 /app/cmd/hemsida/hemsida /opt/lorudden
COPY --from=builder --chown=1001 /app/assets /opt/lorudden/assets

RUN chown 1001 /opt/lorudden

EXPOSE 8080
USER 1001

ENTRYPOINT ["/opt/lorudden/hemsida"]
