// referenced https://medium.com/@juliorenner123/k8s-creating-a-kube-scheduler-plugin-8a826c486a1

package shmscoring

import (
	"context"
	"fmt"
	"strconv"

	"github.com/freckie/shmsched-plugin/apis/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	klog "k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type ShmScoring struct {
	handle framework.Handle
	conn   *ShmmConnector
}

const Name = "ShmScoring"

var _ = framework.ScorePlugin(&ShmScoring{})

func New(obj runtime.Object, h framework.Handle) (framework.Plugin, error) {
	args, ok := obj.(*config.ShmScoringArgs)
	if !ok {
		return nil, fmt.Errorf("[ShmScoring] want args to be of type ShmScoringArgs, got %T", obj)
	}

	klog.Infof("[ShmScoring] args received. args: %v", args)

	return &ShmScoring{
		handle: h,
		conn:   NewShmmConnector(args.Targets),
	}, nil
}

func (s *ShmScoring) Name() string {
	return Name
}

func (s *ShmScoring) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (int64, *framework.Status) {
	var score int64
	klog.Infof("[ShmScoring] Started to scoring nodes")

	// getting pod labels
	labels := p.ObjectMeta.Labels
	modelName, ok := labels["shmfaas-model-name"]
	if !ok {
		return 0, framework.NewStatus(framework.Error, "label \"shmfaas-model-name\" is not specified.")
	}
	tagName, ok := labels["shmfaas-tag-name"]
	if !ok {
		return 0, framework.NewStatus(framework.Error, "label \"shmfaas-tag-name\" is not specified.")
	}
	_shmRequest, ok := labels["shmfaas-shm-request"]
	if !ok {
		return 0, framework.NewStatus(framework.Error, "label \"shmfaas-shm-request\" is not specified.")
	}
	shmRequest, err := strconv.ParseInt(_shmRequest, 10, 64)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, "expected int64 value for \"shmfaas-shm-request\".")
	}
	modelTag := modelName + ":" + tagName

	// getting node metric
	metric, err := s.conn.GetNodeMetric(nodeName)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("error getting node metrics: %s", err))
	}
	hasDeployed := false
	for _, model := range metric.Models {
		if modelTag == model {
			hasDeployed = true
			break
		}
	}

	// scoring
	if !hasDeployed && (shmRequest < metric.ShmDiskFree) {
		score = 1
	} else {
		score = 100 - int64((metric.ShmDiskUsed+shmRequest)*100/metric.ShmDiskAll)
		if score <= 0 {
			score = 1
		}
	}

	klog.Infof("[ShmScoring] node \"%s\" score \"%d\"", nodeName, score)
	return score, nil
}

func (s *ShmScoring) ScoreExtensions() framework.ScoreExtensions {
	return s
}

func (s *ShmScoring) NormalizeScore(
	ctx context.Context,
	state *framework.CycleState,
	pod *v1.Pod,
	scores framework.NodeScoreList,
) *framework.Status {
	var highestScore int64 = 0
	var lowestScore int64 = 100
	for _, node := range scores {
		if highestScore < node.Score {
			highestScore = node.Score
		}
		if lowestScore > node.Score {
			lowestScore = node.Score
		}
	}

	for i, node := range scores {
		scores[i].Score = framework.MaxNodeScore - (node.Score * framework.MaxNodeScore / highestScore)
	}

	klog.Infof("[ShmScoring] final scores: %v", scores)
	return nil
}
