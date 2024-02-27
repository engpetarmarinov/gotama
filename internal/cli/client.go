package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/engpetarmarinov/gotama/internal/base"
	"github.com/engpetarmarinov/gotama/internal/task"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var (
	client = &http.Client{
		Timeout: 5 * time.Second,
	}
)

// TODO: move this to config file, gotama-cli config to generate one?
const baseUrl = "http://localhost:8080/api/v1/"

func get(uri string, params url.Values) (*base.Response, error) {
	apiURL := uri + "?" + params.Encode()

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response base.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func GetTasks(offset int, limit int) ([]task.Response, error) {
	uri := fmt.Sprintf("%stasks", baseUrl)

	params := url.Values{"offset": []string{strconv.Itoa(offset)}, "limit": []string{strconv.Itoa(limit)}}
	rsp, err := get(uri, params)
	if err != nil {
		return nil, err
	}

	if rsp.Error != nil {
		return nil, errors.New(fmt.Sprintf("error received: code: %s, message: %s", rsp.Error.Code, rsp.Error.Message))
	}

	data, ok := rsp.Data.(map[string]interface{})
	if !ok {
		return nil, err
	}

	var tasks []task.Response
	tasksBytes, err := json.Marshal(data["tasks"])
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(tasksBytes, &tasks)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}
