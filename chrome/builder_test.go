package chrome

import (
	"encoding/json"
	"os"
	"testing"
)

func TestAddOperation(t *testing.T) {
	commandNameArray := [8]string{
		"open_web_page",
		"full_page_screenshot",
		"element_screenshot",
		"collect_nodes",
		"click",
		"save_html",
		"sleep",
		"iterate_html",
	}

	byteStringArr := [8]string{
		`{"url": "https://bench-ai.com"}`,
		`{"quality": 10, "snapshot_name": "s1", "name": "test.jpg"}`,
		`{"scale": 10, "snapshot_name": "s1", "name": "test.jpg", "selector": "xpath/123/tada"}`,
		`{"selector": "body", "snapshot_name": "s1"}`,
		`{"selector": "data", "query_type": "search"}`,
		`{"snapshot_name": "s1", "selector": "body"}`,
		`{"ms": 100000}`,
		`{"iter_limit": 10, "pause_time": 30, "starting_snapshot": 1, "image_quality": 10,  "snapshot_name": "s1"}`,
	}

	job := InitFileJob()

	for index, commandName := range commandNameArray {

		err, sess := tempDir()

		if err != nil {
			t.Fatal(err)
		}

		dataMap := map[string]interface{}{}
		dataBytes := []byte(byteStringArr[index])

		if err := json.Unmarshal(dataBytes, &dataMap); err != nil {
			if err := os.RemoveAll(sess); err != nil {
				t.Fatal(err)
			}
			t.Fatal(err)
		}

		_, err = AddOperation(dataMap, commandName, sess, job)

		if err := os.RemoveAll(sess); err != nil {
			t.Fatal(err)
		}

		if err != nil {
			t.Fatal(err)
		}
	}

	_, err := AddOperation(map[string]interface{}{}, "blank", "sess", job)

	if err == nil {
		t.Fatal("did not catch invalid command")
	}
}
