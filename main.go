package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/YuriiChmil/golambda"
)


func main() {
	lambda.Start(golambda.HandleRequest())
}
