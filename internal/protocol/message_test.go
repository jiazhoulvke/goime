package protocol

import (
	"encoding/json"
	"testing"
)

func TestRequestMarshal(t *testing.T) {
	req := Request{Method: "input", Key: "u"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var got Request
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Method != "input" || got.Key != "u" {
		t.Errorf("roundtrip failed: %+v", got)
	}
}

func TestResponseMarshal(t *testing.T) {
	resp := Response{
		Type: "candidates",
		Candidates: &Candidates{
			List: []Candidate{{Text: "输入", Code: "uuru", Weight: 100}},
			Page: 0, Total: 1,
		},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var got Response
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "candidates" || len(got.Candidates.List) != 1 {
		t.Errorf("roundtrip failed: %+v", got)
	}
}
