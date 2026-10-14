//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

const (
	baseURL  = "http://127.0.0.1:3000"
	loginURL = baseURL + "/api/v1/auth/login"
	surveyID = "69aed30b39cb631f8a037458"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

type questionResponse struct {
	QuestionID int `json:"question_id"`
	Score      int `json:"score"`
}

type peerEvalRequest struct {
	EvaluatorID string             `json:"evaluator_id"`
	EvaluatorId int                `json:"evaluatorId"`
	EvaluatedID string             `json:"evaluated_id"`
	EvaluatedId int                `json:"evaluatedId"`
	GroupID     string             `json:"group_id"`
	GroupId     string             `json:"groupId"`
	Responses   []questionResponse `json:"responses"`
}

type group struct {
	id         string
	name       string
	studentIDs []int
}

func main() {
	token, err := login("nicolas.rojas.g@usach.cl", "Hola123.")
	if err != nil {
		fmt.Printf("ERROR login: %v\n", err)
		return
	}
	fmt.Println("Login OK")

	groups := []group{
		{id: "69aed196f5ec209c2c75884f", name: "Grupo 1", studentIDs: []int{47, 46, 56, 39}},
		{id: "69aed196f5ec209c2c758853", name: "Grupo 2", studentIDs: []int{40, 45, 44, 57}},
		{id: "69aed196f5ec209c2c758857", name: "Grupo 3", studentIDs: []int{60, 54, 50, 38}},
		{id: "69aed196f5ec209c2c75885b", name: "Grupo 4", studentIDs: []int{52, 42, 41, 59}},
		{id: "69aed196f5ec209c2c75885f", name: "Grupo 5", studentIDs: []int{61, 49, 53, 51}},
		{id: "69aed196f5ec209c2c758863", name: "Grupo 6", studentIDs: []int{43, 55, 63, 48}},
		{id: "69aed196f5ec209c2c758867", name: "Grupo 7", studentIDs: []int{58, 62}},
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	surveyURL := fmt.Sprintf("%s/api/v1/core/peer-evaluation/surveys/%s/responses", baseURL, surveyID)

	total := 0
	errors := 0

	for _, g := range groups {
		fmt.Printf("\n--- %s (%d estudiantes) ---\n", g.name, len(g.studentIDs))

		for _, evaluatorID := range g.studentIDs {
			for _, evaluatedID := range g.studentIDs {
				responses := buildResponses(rng)

				payload := peerEvalRequest{
					EvaluatorID: fmt.Sprintf("%d", evaluatorID),
					EvaluatorId: evaluatorID,
					EvaluatedID: fmt.Sprintf("%d", evaluatedID),
					EvaluatedId: evaluatedID,
					GroupID:     g.id,
					GroupId:     g.id,
					Responses:   responses,
				}

				err := submitEvaluation(surveyURL, token, payload)
				total++
				if err != nil {
					fmt.Printf("  [ERROR] evaluator=%d -> evaluated=%d: %v\n", evaluatorID, evaluatedID, err)
					errors++
				} else {
					fmt.Printf("  [OK] evaluator=%d -> evaluated=%d\n", evaluatorID, evaluatedID)
				}

				time.Sleep(30 * time.Millisecond)
			}
		}
	}

	fmt.Printf("\n=== Resumen: %d enviados, %d errores ===\n", total, errors)
}

func buildResponses(rng *rand.Rand) []questionResponse {
	responses := make([]questionResponse, 5)
	for i := 0; i < 5; i++ {
		score := rng.Intn(4) + 1
		responses[i] = questionResponse{
			QuestionID: i + 1,
			Score:      score,
		}
	}
	return responses
}

func login(email, password string) (string, error) {
	body, err := json.Marshal(loginRequest{Email: email, Password: password})
	if err != nil {
		return "", fmt.Errorf("marshal login: %w", err)
	}

	resp, err := http.Post(loginURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("post login: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read login response: %w", err)
	}

	var lr loginResponse
	if err := json.Unmarshal(raw, &lr); err != nil {
		return "", fmt.Errorf("unmarshal login response: %w", err)
	}

	if !lr.Success || lr.Data.Token == "" {
		return "", fmt.Errorf("login failed: %s", string(raw))
	}

	return lr.Data.Token, nil
}

func submitEvaluation(url, token string, payload peerEvalRequest) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(raw))
	}

	return nil
}
