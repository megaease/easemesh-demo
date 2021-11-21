package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestMarshal(t *testing.T) {
	buff, _ := json.Marshal(&OrderResponse{})
	fmt.Printf("%s\n", buff)
}
