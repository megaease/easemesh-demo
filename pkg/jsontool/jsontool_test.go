package jsontool

import (
	"testing"
)

var singleObjectData = `{
	"traceId": "865f4e4cee8f8ba1",
	"id": "865f4e4cee8f8ba1",
	"kind": "CLIENT",
	"name": "get",
	"timestamp": 1635177352443025,
	"duration": 49694,
	"localEndpoint": {
	  "serviceName": "award-mesh",
	  "ipv4": "10.244.0.184"
	},
	"tags": {
	  "http.method": "GET",
	  "http.path": "http://localhost:13009/v1/catalog/services?wait=2s&token=",
	  "i": "award-mesh-6fbf7964f4-8gh9l"
	}
      }`

var multipleObjectsData = `[
  {
    "traceId": "865f4e4cee8f8ba1",
    "id": "865f4e4cee8f8ba1",
    "kind": "CLIENT",
    "name": "get",
    "timestamp": 1635177352443025,
    "duration": 49694,
    "localEndpoint": {
      "serviceName": "award-mesh",
      "ipv4": "10.244.0.184"
    },
    "tags": {
      "http.method": "GET",
      "http.path": "http://localhost:13009/v1/catalog/services?wait=2s&token=",
      "i": "award-mesh-6fbf7964f4-8gh9l"
    }
  },
  {
    "traceId": "481fdb7bf192dd6b",
    "parentId": "481fdb7bf192dd6b",
    "id": "4dc21153c7c758a4",
    "kind": "CLIENT",
    "name": "post",
    "timestamp": 1635177352984712,
    "duration": 52210,
    "localEndpoint": {
      "serviceName": "award-mesh",
      "ipv4": "10.244.0.184"
    },
    "tags": {
      "http.method": "POST",
      "http.path": "http://delivery-mesh",
      "i": "award-mesh-6fbf7964f4-8gh9l"
    }
  },
  {
    "traceId": "481fdb7bf192dd6b",
    "id": "481fdb7bf192dd6b",
    "kind": "SERVER",
    "name": "post",
    "timestamp": 1635177352962109,
    "duration": 82538,
    "localEndpoint": {
      "serviceName": "award-mesh",
      "ipv4": "10.244.0.184"
    },
    "remoteEndpoint": {
      "ipv4": "127.0.0.1",
      "port": 58436
    },
    "tags": {
      "http.method": "POST",
      "http.path": "/",
      "http.route": "/",
      "i": "award-mesh-6fbf7964f4-8gh9l"
    },
    "shared": true
  }
]`

func TestUnmarshalObjects(t *testing.T) {
	_, err := UnmarshalObjects([]byte(singleObjectData))
	if err != nil {
		t.Fatalf("unmarshal single object failed: %v", err)
	}

	_, err = UnmarshalObjects([]byte(multipleObjectsData))
	if err != nil {
		t.Fatalf("unmarshal multiple objects failed: %v", err)
	}
}

func TestGetObjects(t *testing.T) {
	objects := GetObjects([]byte(singleObjectData), nil)
	if len(objects) != 1 {
		t.Fatalf("want 1 elements, got %d", len(objects))
	}

	objects = GetObjects([]byte(multipleObjectsData), nil)
	if len(objects) != 3 {
		t.Fatalf("want 3 elements, got %d", len(objects))
	}
}
