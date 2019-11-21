package main

import (
	"github.com/YuriiChmil/golambda/golambda"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(golambda.HandleRequest)
}
