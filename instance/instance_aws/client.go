package instance_aws

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type AWSInstance struct {
	Id         string    `json:"-"`
	Type       string    `json:"type"`
	LaunchTime time.Time `json:"launchTime"`
	PublicIp   string    `json:"publicIp"`
	State      string    `json:"state"`
}

type AWSClient struct {
	ec2 *ec2.Client
	ssm *ssm.Client
}

type AWSInstanceType string

const (
	InstanceTypeSmall  AWSInstanceType = "t3.small"
	InstanceTypeMedium AWSInstanceType = "t3.medium"
	InstanceTypeLarge  AWSInstanceType = "t3.large"
)

// NewClient initializes a new instance of Client
func NewAWSClient() *AWSClient {
	cfg, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		panic("unable to load SDK config")
	}

	// If in development mode, use localstack endpoint for aws services
	if os.Getenv("APP_ENV") != "production" {
		cfg.BaseEndpoint = aws.String(os.Getenv("LOCALSTACK_ENDPOINT"))
	}

	return &AWSClient{
		ec2: ec2.NewFromConfig(cfg),
		ssm: ssm.NewFromConfig(cfg),
	}
}

func (c *AWSClient) CreateInstance(ctx context.Context, instanceType *AWSInstanceType) (*AWSInstance, error) {
	imageId := os.Getenv("AWS_IMAGE_ID_BASE")
	keyName := os.Getenv("AWS_KEY_PAIR_NAME")

	// Read the server-setup script file
	data, err := os.ReadFile("./scripts/instance-setup.sh")
	if err != nil {
		return &AWSInstance{}, fmt.Errorf("failed to read script file: %v", err)
	}

	// Convert the file contents to a string
	encoded := base64.StdEncoding.EncodeToString(data)

	runInstancesInput := &ec2.RunInstancesInput{
		ImageId:      &imageId,
		InstanceType: types.InstanceType(*instanceType),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		KeyName:      &keyName,
		UserData:     aws.String(encoded),
	}

	result, err := c.ec2.RunInstances(ctx, runInstancesInput)
	if err != nil || len(result.Instances) == 0 {
		return &AWSInstance{}, fmt.Errorf("failed to create instance: %v", err)
	}

	instance := result.Instances[0]

	return &AWSInstance{
		Id:         *instance.InstanceId,
		Type:       string(instance.InstanceType),
		LaunchTime: *instance.LaunchTime,
		State:      string(instance.State.Name),
	}, nil
}

func (c *AWSClient) GetInstances(ctx context.Context, instanceIds *[]string) (*[]AWSInstance, error) {
	if len(*instanceIds) == 0 {
		return &[]AWSInstance{}, nil
	}

	result, err := c.ec2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: *instanceIds,
	})
	if err != nil {
		return &[]AWSInstance{}, fmt.Errorf("failed to describe instance: %v", err)
	}

	var instances []AWSInstance

	for _, r := range result.Reservations {
		i := r.Instances[0]

		publicIp := ""

		if i.State.Name == types.InstanceStateNameRunning {
			publicIp = *i.PublicIpAddress
		}

		instances = append(instances, AWSInstance{
			Id:         *i.InstanceId,
			Type:       string(i.InstanceType),
			LaunchTime: *i.LaunchTime,
			State:      string(i.State.Name),
			PublicIp:   publicIp,
		})
	}

	return &instances, nil
}

func (c *AWSClient) GetInstance(ctx context.Context, instanceId *string) (*AWSInstance, error) {
	result, err := c.ec2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{*instanceId},
	})
	if err != nil || len(result.Reservations) == 0 {
		return &AWSInstance{}, fmt.Errorf("failed to describe instance: %v", err)
	}

	instance := result.Reservations[0].Instances[0]

	publicIp := ""
	if instance.State.Name == types.InstanceStateNameRunning {
		publicIp = *instance.PublicIpAddress
	}

	return &AWSInstance{
		Id:         *instance.InstanceId,
		Type:       string(instance.InstanceType),
		LaunchTime: *instance.LaunchTime,
		State:      string(instance.State.Name),
		PublicIp:   publicIp,
	}, nil
}

func (c *AWSClient) UpdateInstance(ctx context.Context, instanceId *string, newInstanceType *AWSInstanceType) (*AWSInstance, error) {
	instance, err := c.GetInstance(ctx, instanceId)
	if err != nil {
		return &AWSInstance{}, err
	}

	// Check if the new instance type is different than the current one
	if instance.Type == string(*newInstanceType) {
		return &AWSInstance{}, errors.New("no changes")
	}

	// Check if the instance is stopped to perform the update
	if instance.State != string(types.InstanceStateNameStopped) {
		return &AWSInstance{}, errors.New("cannot update a running instance")
	}

	_, err = c.ec2.ModifyInstanceAttribute(ctx, &ec2.ModifyInstanceAttributeInput{
		InstanceId: instanceId,
		InstanceType: &types.AttributeValue{
			Value: (*string)(newInstanceType),
		},
	})
	if err != nil {
		return &AWSInstance{}, fmt.Errorf("failed to modify instance type: %v", err)
	}

	instance.Type = string(*newInstanceType)

	return instance, nil
}

func (c *AWSClient) StartInstances(ctx context.Context, instanceIds []string) error {
	_, err := c.ec2.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: instanceIds,
	})
	if err != nil {
		return fmt.Errorf("unable to start instances, %v", err)
	}

	return nil
}

func (c *AWSClient) StopInstances(ctx context.Context, instanceIds []string) error {
	_, err := c.ec2.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: instanceIds,
	})
	if err != nil {
		return fmt.Errorf("unable to stop instances, %v", err)
	}

	return nil
}

func (c *AWSClient) TerminateInstances(ctx context.Context, instanceIds []string) error {
	_, err := c.ec2.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: instanceIds,
	})
	if err != nil {
		return fmt.Errorf("unable to terminate instances, %v", err)
	}

	return nil
}

func (c *AWSClient) GetRunningInstances(ctx context.Context) (*[]string, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running"},
			},
		},
	}

	result, err := c.ec2.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %v", err)
	}

	var instanceIds []string
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			instanceIds = append(instanceIds, *instance.InstanceId)
		}
	}

	return &instanceIds, nil
}

func (c *AWSClient) SendCommand(ctx context.Context, command *string, instanceIds *[]string) error {
	commandInput := &ssm.SendCommandInput{
		InstanceIds:  *instanceIds,
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters: map[string][]string{
			"commands": {*command},
		},
	}

	_, err := c.ssm.SendCommand(ctx, commandInput)
	if err != nil {
		return fmt.Errorf("failed to send command: %v", err)
	}

	return nil
}

func ParseInstanceType(instanceType string) (AWSInstanceType, error) {
	switch instanceType {
	case string(InstanceTypeSmall):
		return InstanceTypeSmall, nil
	case string(InstanceTypeMedium):
		return InstanceTypeMedium, nil
	case string(InstanceTypeLarge):
		return InstanceTypeLarge, nil
	default:
		return "", errors.New("invalid instance type")
	}
}
