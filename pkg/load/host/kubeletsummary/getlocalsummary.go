package kubeletsummary

import (
	"net/http"
	"io/ioutil"
	"encoding/json"

)

func GetLocalSummary() (*Summary, error) {
	resp, err := http.Get("http://localhost:10255/stats/latest_summary")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	summary := &Summary{}
	err = json.Unmarshal(body, summary)
	if err != nil {
		return nil, err
	}

	return summary, nil
}
