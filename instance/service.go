package instance

import (
	"context"
	"fmt"
	"os"
)

// Service provides instance management functionality
type Service struct {
	client *Client
}

// NewService initializes a new instance of Service
func NewService(client *Client) *Service {
	return &Service{client: client}
}

// UpdateAllRunningInstances updates the API on all running instances
func (s *Service) UpdateAllRunningInstanceAPIs(ctx context.Context) error {
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
