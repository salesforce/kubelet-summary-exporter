/*
 * Copyright (c) 2022, salesforce.com, inc.
 * All rights reserved.
 * SPDX-License-Identifier: BSD-3-Clause
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
 */
package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Scraper struct {
	tokenPath string
	timeout   time.Duration
	targetIP  string
	storage   *prometheus.Desc
	errors    *prometheus.Desc
	errCnt    float64
	logger    *zap.Logger
}

func NewScraper(logger *zap.Logger, targetIP string, tokenPath string, timeout time.Duration) *Scraper {
	return &Scraper{
		tokenPath: tokenPath,
		timeout:   timeout,
		targetIP:  targetIP,
		logger:    logger.With(zap.String("component", "scraper")),
		storage: prometheus.NewDesc(
			prometheus.BuildFQName("kube_pod", "", "ephemeral_storage_used_bytes"),
			"Ephemeral storage used in bytes",
			[]string{"node", "namespace", "pod"},
			nil),
		errors: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary_exporter", "", "errors"),
			"Errors scraping kubelet stats summary",
			[]string{"type"},
			nil),
	}
}

func (s *Scraper) Describe(ch chan<- *prometheus.Desc) {
	ch <- s.storage
	ch <- s.errors
}

func (s *Scraper) Collect(ch chan<- prometheus.Metric) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s:10250/stats/summary", s.targetIP), nil)
	if err != nil {
		s.logger.Error("failed to create request", zap.Error(err))
		return
	}

	token, err := os.ReadFile(s.tokenPath)
	if err != nil {
		s.logger.Fatal("unable to load specified token", zap.String("file", s.tokenPath), zap.Error(err))
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	client := &http.Client{
		Timeout: s.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		s.errCnt++
		ch <- prometheus.MustNewConstMetric(
			s.errors,
			prometheus.CounterValue,
			s.errCnt,
			"request error",
		)
		s.logger.Warn("failed to make request to stats/summary", zap.Error(err))
		return
	}

	if resp.StatusCode != http.StatusOK {
		s.errCnt++
		ch <- prometheus.MustNewConstMetric(
			s.errors,
			prometheus.CounterValue,
			s.errCnt,
			"status error",
		)
		s.logger.Warn("got unexpected status for stats/summary", zap.String("status", resp.Status))
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.errCnt++
		ch <- prometheus.MustNewConstMetric(
			s.errors,
			prometheus.CounterValue,
			s.errCnt,
			"read body error",
		)
		s.logger.Error("failed to read body", zap.Error(err))
		return
	}

	summary, err := s.parse(body)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			s.errors,
			prometheus.CounterValue,
			s.errCnt,
			"parse body error",
		)
		s.logger.Error("failed to parse body", zap.Error(err))
		return
	}

	for _, pod := range summary.Pods {
		if pod == nil {
			continue
		}
		if pod.EphemeralStorage != nil {
			usedBytes := pod.EphemeralStorage.UsedBytes
			ch <- prometheus.MustNewConstMetric(
				s.storage,
				prometheus.GaugeValue,
				float64(usedBytes),
				summary.Node.NodeName, pod.PodRef.Namespace, pod.PodRef.Name, // node, namespace, pod
			)
		}
	}
}

func (s *Scraper) parse(body []byte) (*Summary, error) {
	var summary Summary
	err := json.Unmarshal(body, &summary)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

// Summary represents the response summary endpoint
type Summary struct {
	Node Node   `json:"node"`
	Pods []*Pod `json:"pods"`
}

// Node represents the node from the summary endpoint
type Node struct {
	NodeName string `json:"nodeName"`
}

// Pod represents pod spec in the summary endpoint
type Pod struct {
	/*
		EXAMPLE:
		"podRef": {
		     "name": "configs-service-59c9c7586b-5jchj",
		     "namespace": "onprem",
		     "uid": "5fbb63da-d0a3-4493-8d27-6576b63119f5"
		    }
	*/
	PodRef struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"podRef"`
	/*
		  Don't parse the list of volumes
			EXAMPLE:
			"volume": [
			     {...},
			     {...}
			    ]
	*/
	//Volumes    []*Volume `json:"volume"`
	EphemeralStorage *Volume `json:"ephemeral-storage"`
}

// Volume represents the volume struct
/*
EXAMPLE:
{
"time": "2019-11-25T20:33:19Z",
"availableBytes": 25674719232,
"capacityBytes": 25674731520,
"usedBytes": 12288,
"inodesFree": 6268236,
"inodes": 6268245,
"inodesUsed": 9,
"name": "vault-client"
}
*/
// https://github.com/kubernetes/kubernetes/blob/v1.18.5/pkg/volume/volume.go
// https://github.com/kubernetes/kubernetes/blob/v1.18.5/pkg/volume/csi/csi_client.go#L553
type Volume struct {
	// The time at which these stats were updated.
	//	Time metav1.Time `json:"time"`

	// Used represents the total bytes used by the Volume.
	// Note: For block devices this maybe more than the total size of the files.
	UsedBytes int64 `json:"usedBytes"` // TODO: use uint64 here as well?

	// Capacity represents the total capacity (bytes) of the volume's
	// underlying storage. For Volumes that share a filesystem with the host
	// (e.g. emptydir, hostpath) this is the size of the underlying storage,
	// and will not equal Used + Available as the fs is shared.
	//CapacityBytes int64 `json:"capacityBytes"`

	// Available represents the storage space available (bytes) for the
	// Volume. For Volumes that share a filesystem with the host (e.g.
	// emptydir, hostpath), this is the available space on the underlying
	// storage, and is shared with host processes and other Volumes.
	AvailableBytes int64 `json:"availableBytes"`

	// InodesUsed represents the total inodes used by the Volume.
	//InodesUsed uint64 `json:"inodesUsed"`

	// Inodes represents the total number of inodes available in the volume.
	// For volumes that share a filesystem with the host (e.g. emptydir, hostpath),
	// this is the inodes available in the underlying storage,
	// and will not equal InodesUsed + InodesFree as the fs is shared.
	//Inodes uint64 `json:"inodes"`

	// InodesFree represent the inodes available for the volume.  For Volumes that share
	// a filesystem with the host (e.g. emptydir, hostpath), this is the free inodes
	// on the underlying storage, and is shared with host processes and other volumes
	//InodesFree uint64 `json:"inodesFree"`
	/*
		Name   string `json:"name"`

		PvcRef struct {
			PvcName      string `json:"name"`
			PvcNamespace string `json:"namespace"`
		} `json:"pvcRef"`
	*/
}

func (p *Pod) MarshalLogObject(oe zapcore.ObjectEncoder) error {
	if p != nil {
		oe.AddString("name", p.PodRef.Name)
		oe.AddString("namespace", p.PodRef.Name)
	}
	return nil
}
