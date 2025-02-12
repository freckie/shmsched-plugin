package shmscoring

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/freckie/shmsched-plugin/apis/config"
	klog "k8s.io/klog/v2"
)

const (
	urlFormat string = "http://%s:%s%s" // {0}=ip, {1}=port, {2}=endpoint
)

type ShmmConnector struct {
	Targets map[string]config.ShmScoringTarget
}

func NewShmmConnector(targets []config.ShmScoringTarget) *ShmmConnector {
	s := &ShmmConnector{
		Targets: make(map[string]config.ShmScoringTarget),
	}
	for _, t := range targets {
		s.Targets[t.NodeName] = t
	}

	klog.Infof("[ShmScoring] Targets in ShmmConnector : %v", s.Targets)

	return s
}

func (s *ShmmConnector) GetNodeMetric(nodeName string) (NodeMetric, error) {
	result := NodeMetric{}

	t, ok := s.Targets[nodeName]
	if !ok {
		return result, fmt.Errorf("nodeName \"%s\" not found in ShmmConnector.", nodeName)
	}

	url := fmt.Sprintf(urlFormat, t.IP, t.Port, "/metrics/mem")
	resp, err := http.Get(url)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	var jsonResult struct {
		Data metricsResp `json:"data"`
	}
	if err = json.Unmarshal(body, &jsonResult); err != nil {
		return result, err
	}

	result.ShmSum = jsonResult.Data.ShmSum
	result.ShmDiskUsed = jsonResult.Data.ShmDisk.Used
	result.ShmDiskFree = jsonResult.Data.ShmDisk.Free
	result.ShmDiskAll = jsonResult.Data.ShmDisk.All
	result.ModelCount = jsonResult.Data.ModelCount
	result.Models = make([]string, result.ModelCount)
	for i, m := range jsonResult.Data.Models {
		result.Models[i] = fmt.Sprintf("%s:%s", m.ModelName, m.TagName)
	}

	return result, nil
}
