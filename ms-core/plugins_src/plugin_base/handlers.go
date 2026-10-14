package main

import (
	"context"
	"fmt"
)

func (p *MyPlugin) createTest(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("create_test not implemented")
}

func (p *MyPlugin) getTests(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("get_tests not implemented")
}

func (p *MyPlugin) getTest(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("get_test not implemented")
}

func (p *MyPlugin) updateTest(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("update_test not implemented")
}

func (p *MyPlugin) deleteTest(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("delete_test not implemented")
}
