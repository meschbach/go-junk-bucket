package files

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type ExampleJSONFile struct {
	KeyA string `json:"key-a,omitempty"`
}

func TestParseJSONFile(t *testing.T)  {
	var example ExampleJSONFile
	err := ParseJSONFile("jsonfile-example.json", &example)
	if err != nil {
		panic(err)
	}

	assert.Equal(t,"an example value", example.KeyA)
}
