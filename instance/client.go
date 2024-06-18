package instance

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
	"github.com/mooncorn/gshub-main-api/config"
)

type Instance struct {
	Id         string    `json:"-"`
	Type       string    `json:"type"`
	LaunchTime time.Time `json:"launchTime"`
	PublicIp   string    `json:"publicIp"`
	State      string    `json:"state"`
}

type Client struct {
	ec2 *ec2.Client
	ssm *ssm.Client
}

type InstanceType string

const (
	InstanceTypeSmall  InstanceType = "t3.small"
	InstanceTypeMedium InstanceType = "t3.medium"
	InstanceTypeLarge  InstanceType = "t3.large"
)

// NewClient initializes a new instance of Client
func NewClient() *Client {
	cfg, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		panic("unable to load SDK config")
	}

	// If in development mode, use localstack endpoint for aws services
	if config.Env.AppEnv != "production" {
		cfg.BaseEndpoint = aws.String(config.Env.LocalStackEndpoint)
	}

	return &Client{
		ec2: ec2.NewFromConfig(cfg),
		ssm: ssm.NewFromConfig(cfg),
	}
}

func (c *Client) CreateInstance(ctx context.Context, instanceType *InstanceType) (*Instance, error) {
	imageId := config.Env.AWSImageIdBase
	keyName := config.Env.AWSKeyPairName

	// Read the server-setup script file
	data, err := os.ReadFile("./scripts/instance-setup.sh")
	if err != nil {
		return &Instance{}, fmt.Errorf("failed to read script file: %v", err)
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
		return &Instance{}, fmt.Errorf("failed to create instance: %v", err)
	}

	instance := result.Instances[0]

	return &Instance{
		Id:         *instance.InstanceId,
		Type:       string(instance.InstanceType),
		LaunchTime: *instance.LaunchTime,
		State:      string(instance.State.Name),
	}, nil
}

func (c *Client) GetInstances(ctx context.Context, instanceIds *[]string) (*[]Instance, error) {
	if len(*instanceIds) == 0 {
		return &[]Instance{}, nil
	}

	result, err := c.ec2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: *instanceIds,
	})
	if err != nil {
		return &[]Instance{}, fmt.Errorf("failed to describe instance: %v", err)
	}

	var instances []Instance

	for _, r := range result.Reservations {
		i := r.Instances[0]

		publicIp := ""

		if i.State.Name == types.InstanceStateNameRunning {
			publicIp = *i.PublicIpAddress
		}

		instances = append(instances, Instance{
			Id:         *i.InstanceId,
			Type:       string(i.InstanceType),
			LaunchTime: *i.LaunchTime,
			State:      string(i.State.Name),
			PublicIp:   publicIp,
		})
	}

	return &instances, nil
}

func (c *Client) GetInstance(ctx context.Context, instanceId *string) (*Instance, error) {
	result, err := c.ec2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{*instanceId},
	})
	if err != nil || len(result.Reservations) == 0 {
		return &Instance{}, fmt.Errorf("failed to describe instance: %v", err)
	}

	instance := result.Reservations[0].Instances[0]

	publicIp := ""
	if instance.State.Name == types.InstanceStateNameRunning {
		publicIp = *instance.PublicIpAddress
	}

	return &Instance{
		Id:         *instance.InstanceId,
		Type:       string(instance.InstanceType),
		LaunchTime: *instance.LaunchTime,
		State:      string(instance.State.Name),
		PublicIp:   publicIp,
	}, nil
}

func (c *Client) UpdateInstance(ctx context.Context, instanceId *string, newInstanceType *InstanceType) (*Instance, error) {
	instance, err := c.GetInstance(ctx, instanceId)
	if err != nil {
		return &Instance{}, err
	}

	// Check if the new instance type is different than the current one
	if instance.Type == string(*newInstanceType) {
		return &Instance{}, errors.New("no changes")
	}

	// Check if the instance is stopped to perform the update
	if instance.State != string(types.InstanceStateNameStopped) {
		return &Instance{}, errors.New("cannot update a running instance")
	}

	_, err = c.ec2.ModifyInstanceAttribute(ctx, &ec2.ModifyInstanceAttributeInput{
		InstanceId: instanceId,
		InstanceType: &types.AttributeValue{
			Value: (*string)(newInstanceType),
		},
	})
	if err != nil {
		return &Instance{}, fmt.Errorf("failed to modify instance type: %v", err)
	}

	instance.Type = string(*newInstanceType)

	return instance, nil
}

func (c *Client) StartInstances(ctx context.Context, instanceIds []string) error {
	_, err := c.ec2.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: instanceIds,
	})
	if err != nil {
		return fmt.Errorf("unable to start instances, %v", err)
	}

	return nil
}

func (c *Client) StopInstances(ctx context.Context, instanceIds []string) error {
	_, err := c.ec2.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: instanceIds,
	})
	if err != nil {
		return fmt.Errorf("unable to stop instances, %v", err)
	}

	return nil
}

func (c *Client) TerminateInstances(ctx context.Context, instanceIds []string) error {
	_, err := c.ec2.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: instanceIds,
	})
	if err != nil {
		return fmt.Errorf("unable to terminate instances, %v", err)
	}

	return nil
}

func (c *Client) GetRunningInstances(ctx context.Context) (*[]string, error) {
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

func (c *Client) SendCommand(ctx context.Context, command *string, instanceIds *[]string) error {
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

func ParseInstanceType(instanceType string) (InstanceType, error) {
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
