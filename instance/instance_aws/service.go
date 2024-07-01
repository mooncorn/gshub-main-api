package instance_aws

import (
	"context"
	"fmt"
	"os"
)

// Service provides instance management functionality
type AWSService struct {
	client *AWSClient
}

// NewService initializes a new instance of Service
func NewService(client *AWSClient) *AWSService {
	return &AWSService{client: client}
}

// UpdateAllRunningInstances updates the API on all running instances
func (s *AWSService) UpdateAllRunningInstanceAPIs(ctx context.Context) error {
	instanceIds, err := s.client.GetRunningInstances(ctx)
	if err != nil {
		return fmt.Errorf("failed to get running instances: %v", err)
	}

	if len(*instanceIds) == 0 {
		return nil
	}

	scriptPath := "./scripts/instance-update.sh"
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("unable to read file: %s", scriptPath)
	}

	command := string(data)

	err = s.client.SendCommand(ctx, &command, instanceIds)
	if err != nil {
		return fmt.Errorf("failed to update instance apis: %v", err)
	}

	return nil
}
