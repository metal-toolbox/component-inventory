FROM golang:1.21.3-alpine3.18 AS build

ARG APP_NAME
WORKDIR "/go/src/github.com/metal-toolbox/${APP_NAME}"
COPY go.mod go.sum ./
RUN go mod download

ARG LDFLAG_LOCATION
ARG GIT_COMMIT
ARG GIT_BRANCH
ARG GIT_SUMMARY
ARG VERSION
ARG BUILD_DATE

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "/${APP_NAME}" \
-ldflags \
"-X ${LDFLAG_LOCATION}.GitCommit=${GIT_COMMIT} \
-X ${LDFLAG_LOCATION}.GitBranch=${GIT_BRANCH} \
-X ${LDFLAG_LOCATION}.GitSummary=${GIT_SUMMARY} \
-X ${LDFLAG_LOCATION}.AppVersion=${VERSION} \
-X ${LDFLAG_LOCATION}.BuildDate=${BUILD_DATE}"

FROM alpine:3.19.1
ARG APP_NAME
ENV APP_NAME=${APP_NAME}

RUN echo "${APP_NAME}"
RUN apk -U add curl

COPY --from=build "${APP_NAME}" /fleet/${APP_NAME}

RUN adduser -D fleet
RUN chown -R fleet:fleet /fleet

USER fleet

ENTRYPOINT "/fleet/${APP_NAME}" "$0" "$@"
