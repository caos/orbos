package api

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func Call(region, secretID, accessKey string) error {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials("ORBITER", secretID, accessKey),
	})
	if err != nil {
		return err
	}

	svc := ec2.New(sess)

	svc.RunInstances()

	return nil
}
