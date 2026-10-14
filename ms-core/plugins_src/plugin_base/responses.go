package main

import (
	"context"
	"fmt"
)

func (p *MyPlugin) submitResponse(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("submit_response not implemented")
}

func (p *MyPlugin) getResponses(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("get_responses not implemented")
}

func (p *MyPlugin) getStudentResponse(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("get_student_response not implemented")
}

func (p *MyPlugin) getSectionResponses(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("get_section_responses not implemented")
}
