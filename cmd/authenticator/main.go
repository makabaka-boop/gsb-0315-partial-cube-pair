package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"cube-auth/internal/cube"
)

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr))
}

func run(stdin io.Reader, stdout, stderr io.Writer) int {
	data, err := io.ReadAll(stdin)
	if err != nil {
		writeError(stderr, "cannot read input: "+err.Error())
		return 2
	}

	var faces cube.Faces
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&faces); err != nil {
		writeError(stderr, "invalid JSON: "+err.Error())
		return 2
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); err == nil {
		writeError(stderr, "invalid JSON: expected exactly one JSON object")
		return 2
	} else if err != io.EOF {
		writeError(stderr, "invalid JSON: "+err.Error())
		return 2
	}
	if err := checkShape(faces); err != nil {
		writeError(stderr, "invalid request: "+err.Error())
		return 2
	}

	report := cube.ValidatePartial(faces)
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		writeError(stderr, "cannot encode report: "+err.Error())
		return 2
	}
	encoded = append(encoded, '\n')
	if _, err := stdout.Write(encoded); err != nil {
		writeError(stderr, "cannot write output: "+err.Error())
		return 2
	}
	if report.Legal {
		return 0
	}
	return 1
}

func checkShape(faces cube.Faces) error {
	expected := []string{"U", "R", "F", "D", "L", "B"}
	if len(faces) != len(expected) {
		return fmt.Errorf("exactly faces %v are required", expected)
	}
	for _, name := range expected {
		face, ok := faces[name]
		if !ok {
			return fmt.Errorf("missing face %q", name)
		}
		if len(face) != 9 {
			return fmt.Errorf("face %q must contain exactly nine stickers", name)
		}
		for i, color := range face {
			if color == "" {
				return fmt.Errorf("face %q sticker %d has an empty color", name, i)
			}
		}
	}
	return nil
}

func writeError(w io.Writer, message string) {
	encoded, _ := json.Marshal(map[string]string{"error": message})
	encoded = append(encoded, '\n')
	_, _ = w.Write(encoded)
}
