package jsontool

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

// UnmarshalObjects unmarshals json object or array of objects.
func UnmarshalObjects(data []byte) ([]map[string]interface{}, error) {
	var v interface{}
	err := jsoniter.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}

	switch v := v.(type) {
	case map[string]interface{}:
		return []map[string]interface{}{v}, nil
	case []interface{}:
		var objects []map[string]interface{}
		for i, object := range v {
			obj, ok := object.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("%d element got wrong format: %v", i+1, object)
			}
			objects = append(objects, obj)
		}
		return objects, nil
	default:
		return nil, fmt.Errorf("unsupport format: %T", v)
	}
}

// GetObjects gets objects from an array or an object.
func GetObjects(data []byte, filter func(object []byte) bool) [][]byte {
	objects := [][]byte{}

	result := gjson.GetBytes(data, "@this")
	for _, res := range result.Array() {
		if !res.IsObject() {
			continue
		}

		if filter != nil && !filter([]byte(res.Raw)) {
			continue
		}

		objects = append(objects, []byte(res.Raw))
	}

	return objects
}
