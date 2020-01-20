package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func main() {

	ctx := context.Background()
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("eu-west-3"), // Paris
		Credentials: credentials.NewCredentials(&credentials.SharedCredentialsProvider{}),
	})
	if err != nil {
		panic(err)
	}

	iamSvc := iam.New(sess)
	userPolicyName := "SSM"
	userPolicyARN := strPtr(fmt.Sprintf("arn:aws:iam::976485266142:policy/%s", userPolicyName))
	userPolicyDoc := `{
		"Version":"2012-10-17",
		"Statement":[{
			"Effect":"Allow",
			"Action":[
			   "ssm:SendCommand"
			],
			"Resource":[
			   "arn:aws:ec2:eu-west-3:976485266142:instance/*"
			],
			"Condition":{
			   "StringLike":{
				  "ssm:resourceTag/Operator":[
					 "Orbiter"
				  ]
			   }
			}
		 }, {
			"Effect":"Allow",
			"Action":[
			   "ssm:SendCommand"
			],
			"Resource":[
			   "arn:aws:ssm:eu-west-3::document/AWS-StartSSHSession"
			]
		 }]
	}`

	if _, err := iamSvc.CreatePolicyWithContext(ctx, &iam.CreatePolicyInput{
		PolicyDocument: strPtr(userPolicyDoc),
		PolicyName:     strPtr(userPolicyName),
	}); err != nil && err.(awserr.Error).Code() != "EntityAlreadyExists" {
		panic(err)
	}

	if _, err := iamSvc.AttachUserPolicyWithContext(ctx, &iam.AttachUserPolicyInput{
		UserName:  strPtr("orbiter"),
		PolicyArn: userPolicyARN,
	}); err != nil && err.(awserr.Error).Code() != "EntityAlreadyExists" {
		panic(err)
	}

	commonName := strPtr("OrbiterInstance")
	assumePolicyDoc := `{
		"Version":"2012-10-17",
		"Statement":[{
			"Effect":"Allow",
			"Principal":{
				"Service":"ssm.amazonaws.com"
			},
			"Action":"sts:AssumeRole"
		}]
	}`
	if _, err := iamSvc.CreateRoleWithContext(ctx, &iam.CreateRoleInput{
		RoleName:                 commonName,
		AssumeRolePolicyDocument: strPtr(assumePolicyDoc),
	}); err != nil && err.(awserr.Error).Code() != "EntityAlreadyExists" {
		panic(err)
	}

	if _, err := iamSvc.CreateInstanceProfileWithContext(ctx, &iam.CreateInstanceProfileInput{
		InstanceProfileName: commonName,
	}); err != nil && err.(awserr.Error).Code() != "EntityAlreadyExists" {
		panic(err)
	}

	if _, err = iamSvc.AddRoleToInstanceProfileWithContext(ctx, &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: commonName,
		RoleName:            commonName,
	}); err != nil && err.(awserr.Error).Code() != "LimitExceeded" {
		panic(err)
	}

	ec2svc := ec2.New(sess)
	runResp, err := ec2svc.RunInstancesWithContext(ctx, &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-0e9e6ba6d3d38faa8"),
		InstanceType: aws.String("t2.micro"),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		TagSpecifications: []*ec2.TagSpecification{{
			ResourceType: strPtr("instance"),
			Tags: []*ec2.Tag{{
				Key:   strPtr("Operator"),
				Value: strPtr("Orbiter"),
			}, {
				Key:   strPtr("Orb"),
				Value: strPtr("orbitertest"),
			}},
		}},
		//		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
		//			Name: commonName,
		//		},
	})
	if err != nil {
		panic(err)
	}

	inst := runResp.Instances[0]
	fmt.Println("Instance ID", *inst.InstanceId)

	var status int64
	for status != 16 {
		time.Sleep(3 * time.Second)
		statusResp, err := ec2svc.DescribeInstanceStatusWithContext(ctx, &ec2.DescribeInstanceStatusInput{
			InstanceIds: []*string{inst.InstanceId},
		})
		if err != nil {
			panic(err)
		}
		if len(statusResp.InstanceStatuses) == 1 {
			status = *statusResp.InstanceStatuses[0].InstanceState.Code
		}
		fmt.Printf("Instance %s is %s\n", *inst.InstanceId, *inst.State.Name)
	}

	ssmSvc := ssm.New(sess)
	_, err = ssmSvc.SendCommandWithContext(ctx, &ssm.SendCommandInput{

		InstanceIds:  []*string{inst.InstanceId},
		DocumentName: strPtr("AWS-StartSSHSession"),
	})
	if err != nil {
		panic(err)
	}
}

func strPtr(str string) *string { return &str }
func int64Ptr(num int64) *int64 { return &num }
