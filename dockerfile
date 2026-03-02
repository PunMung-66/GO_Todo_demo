FROM golang:1.25 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

ENV GOARCH=amd64

#  because in golang image, is a linux image, so we use debian image to run the builded file instaed
#  if we run the builded file in golang image, it will be very big, because it contains all the golang files, but if we run the builded file in debian image, 
#  it will be very small, because it only contains the necessary files to run the application.

RUN 	go build \
		-ldflags "-X main.buildcommit=`git rev-parse --short HEAD` \
		-X main.buildtime=`date "+%Y-%m-%dT%H:%M:%S%z"`" \
		-o app
#  we set the build file is app, and the build time and build commit is set by git and date command, so we can get the build time and build commit in the code.
## Deploy
# go need to use distroless to deploy, because distroless is a minimal image that only contains the necessary files to run the application, and it is more secure than other images.
FROM gcr.io/distroless/base-debian12
# /floder/file to /floder
COPY --from=build /app/app /app

EXPOSE 8081

USER nonroot:nonroot

CMD ["/app"]