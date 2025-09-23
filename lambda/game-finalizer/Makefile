.PHONY: build package deploy clean

# Build the Lambda function
build:
	GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

# Package the Lambda function for deployment
package: build
	zip lambda-function.zip bootstrap

# Clean build artifacts
clean:
	rm -f bootstrap lambda-function.zip

# Build and package
all: clean build package

# Deploy using AWS CLI (optional)
deploy: package
	aws lambda update-function-code --function-name apollo-game-finalizer --zip-file fileb://lambda-function.zip