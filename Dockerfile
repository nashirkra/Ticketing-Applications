FROM golang:1.13.15

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/github.com/nashirkra/Ticketing-Applications

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

# This container exposes port 8090 to the outside world
EXPOSE 8090

# Run the executable
CMD ["Ticketing-Applications"]
