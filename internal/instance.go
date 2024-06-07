package internal

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmTypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/mooncorn/gshub-main-api/config"
)

type Instance struct {
	Id         string    `json:"id"`
	Type       string    `json:"type"`
	LaunchTime time.Time `json:"launchTime"`
	PublicIp   string    `json:"publicIp"`
	State      string    `json:"state"`
	Setup      string    `json:"setup"`
	// Ready      bool      `json:"ready"`
}

type InstanceClient struct {
	ec2 *ec2.Client
}

type ContainerOptions struct {
	Image  string   `json:"image"`  // container image
	Env    []string `json:"env"`    // env vars for example: {"eula=true", "version=latest"}
	Volume string   `json:"volume"` // "path_to_server_data_in_container"
	Ports  []string `json:"ports"`  // ports to bind for example: {"25565:25565", "25566:25570"}
}

func NewInstanceClient(ctx context.Context) (*InstanceClient, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return &InstanceClient{}, fmt.Errorf("failed to load AWS config: %v", err)
	}

	if config.Env.LocalStackEndpoint != "" {
		cfg.BaseEndpoint = aws.String(config.Env.LocalStackEndpoint)
	}

	return &InstanceClient{ec2: ec2.NewFromConfig(cfg)}, nil
}

func (c *InstanceClient) CreateInstance(ctx context.Context, instanceType types.InstanceType) (*Instance, error) {
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
		InstanceType: instanceType,
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

func (c *InstanceClient) GetInstances(ctx context.Context, instanceIds []string) ([]Instance, error) {
	if len(instanceIds) == 0 {
		return []Instance{}, nil
	}

	result, err := c.ec2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: instanceIds,
	})
	if err != nil {
		return []Instance{}, fmt.Errorf("failed to describe instance: %v", err)
	}

	var instances []Instance

	for _, r := range result.Reservations {
		i := r.Instances[0]

		setupState := "unknown"
		publicIp := ""
		// ready := false

		if i.State.Name == types.InstanceStateNameRunning {
			// Set public ip address
			publicIp = *i.PublicIpAddress

			// Check setup state
			setupState, err = c.checkSetupStatus(ctx, *i.InstanceId)
			if err != nil {
				setupState = "error"
			}

			// Check api ready
			// if i.State.Name == types.InstanceStateNameRunning {
			// 	ready, err = c.HealthCheckInstanceAPI(ctx, *i.InstanceId)
			// 	if err != nil {
			// 		ready = false
			// 	}
			// }
		}

		instances = append(instances, Instance{
			Id:         *i.InstanceId,
			Type:       string(i.InstanceType),
			LaunchTime: *i.LaunchTime,
			State:      string(i.State.Name),
			PublicIp:   publicIp,
			Setup:      setupState,
			// Ready:      ready,
		})
	}

	return instances, nil
}

func (c *InstanceClient) GetInstance(ctx context.Context, instanceId string) (*Instance, error) {
	result, err := c.ec2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceId},
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

func (c *InstanceClient) TerminateInstance(ctx context.Context, instanceId string) (*Instance, error) {
	instance, err := c.GetInstance(ctx, instanceId)
	if err != nil {
		return &Instance{}, err
	}

	_, err = c.ec2.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceId},
	})
	if err != nil {
		return &Instance{}, fmt.Errorf("failed to terminate instance: %v", err)
	}

	return instance, nil
}

func (c *InstanceClient) UpdateInstance(ctx context.Context, instanceId string, newInstanceType types.InstanceType) (*Instance, error) {
	instance, err := c.GetInstance(ctx, instanceId)
	if err != nil {
		return &Instance{}, err
	}

	// Check if the new instance type is different than the current one
	if instance.Type == string(newInstanceType) {
		return &Instance{}, errors.New("no changes")
	}

	// Check if the instance is stopped to perform the update
	if instance.State != string(types.InstanceStateNameStopped) {
		return &Instance{}, errors.New("stop the instance to update it")
	}

	_, err = c.ec2.ModifyInstanceAttribute(ctx, &ec2.ModifyInstanceAttributeInput{
		InstanceId: &instanceId,
		InstanceType: &types.AttributeValue{
			Value: (*string)(&newInstanceType),
		},
	})
	if err != nil {
		return &Instance{}, fmt.Errorf("failed to modify instance type: %v", err)
	}

	instance.Type = string(newInstanceType)

	return instance, nil
}

func (c *InstanceClient) StartInstance(ctx context.Context, instanceId string) error {
	_, err := c.ec2.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{instanceId},
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *InstanceClient) StopInstance(ctx context.Context, instanceId string) error {
	_, err := c.ec2.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceId},
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *InstanceClient) HealthCheckInstanceAPI(ctx context.Context, instanceId string) (bool, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return false, fmt.Errorf("unable to load SDK config, %v", err)
	}

	ssmClient := ssm.NewFromConfig(cfg)

	// Make sure SSM is ready by executing checkSetupStatus function
	status, err := c.checkSetupStatus(ctx, instanceId)
	if err != nil {
		return false, fmt.Errorf("unable to check setup status: %v", err)
	}

	if status != "complete" {
		return false, fmt.Errorf("setup is not complete, current status: %v", status)
	}

	// Use SSM to execute a script which pings the API at localhost:3001
	commandInput := &ssm.SendCommandInput{
		InstanceIds:  []string{instanceId},
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters: map[string][]string{
			"commands": {
				"curl -s -o /dev/null -w '%{http_code}' http://localhost:3001/ || echo '0'",
			},
		},
	}

	commandOutput, err := ssmClient.SendCommand(ctx, commandInput)
	if err != nil {
		return false, fmt.Errorf("unable to send command, %v", err)
	}

	commandId := *commandOutput.Command.CommandId

	// Wait for command to execute using ssm.NewCommandExecutedWaiter
	waiter := ssm.NewCommandExecutedWaiter(ssmClient)
	waiterInput := &ssm.GetCommandInvocationInput{
		CommandId:  aws.String(commandId),
		InstanceId: aws.String(instanceId),
	}

	err = waiter.Wait(ctx, waiterInput, time.Minute)
	if err != nil {
		return false, fmt.Errorf("waiting for command execution failed: %v", err)
	}

	// Retrieve command output
	output, err := ssmClient.GetCommandInvocation(ctx, waiterInput)
	if err != nil {
		return false, fmt.Errorf("unable to get command output, %v", err)
	}

	fmt.Println(*output.StandardOutputContent)

	// Return a flag indicating API status
	return strings.TrimSpace(*output.StandardOutputContent) == "200", nil
}

func buildDockerRunCommand(opts *ContainerOptions) string {
	var cmd strings.Builder

	// Start with the basic Docker run command
	cmd.WriteString("sudo docker run -d --name main")

	// Append environment variables
	for _, env := range opts.Env {
		cmd.WriteString(fmt.Sprintf(" -e \"%s\"", env))
	}

	// Append volume if specified
	if opts.Volume != "" {
		cmd.WriteString(fmt.Sprintf(" -v %s", opts.Volume))
	}

	// Append port mappings
	for _, port := range opts.Ports {
		cmd.WriteString(fmt.Sprintf(" -p %s", port))
	}

	// Append the image last
	cmd.WriteString(fmt.Sprintf(" %s", opts.Image))

	return cmd.String()
}

func (c *InstanceClient) checkSetupStatus(ctx context.Context, instanceId string) (string, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config, %v", err)
	}

	ssmClient := ssm.NewFromConfig(cfg)

	commandInput := &ssm.ListCommandInvocationsInput{
		InstanceId: aws.String(instanceId),
		MaxResults: aws.Int32(10),
	}

	commandOutput, err := ssmClient.ListCommandInvocations(ctx, commandInput)
	if err != nil {
		return "", fmt.Errorf("unable to list command invocations, %v", err)
	}

	if len(commandOutput.CommandInvocations) == 0 {
		return "not started", nil
	}

	// Find the most recent command invocation
	var lastCommand *ssmTypes.CommandInvocation
	for _, cmd := range commandOutput.CommandInvocations {
		if lastCommand == nil || cmd.RequestedDateTime.After(*lastCommand.RequestedDateTime) {
			lastCommand = &cmd
		}
	}

	switch lastCommand.Status {
	case "InProgress", "Pending":
		return "in progress", nil
	case "Success":
		return "complete", nil
	case "Failed", "Cancelled", "TimedOut":
		return "failed", nil
	}

	return "unknown", nil
}
