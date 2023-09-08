/*
 * Copyright (c) 2022, salesforce.com, inc.
 * All rights reserved.
 * SPDX-License-Identifier: BSD-3-Clause
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
 */
package scraper

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
)

func TestDecodingJson(t *testing.T) {
	for _, tc := range []struct {
		Name        string
		InputFile   string
		WantSummary *Summary
	}{
		{
			Name:      "validate example one test data",
			InputFile: "testdata/example.yaml",
			WantSummary: &Summary{Node: Node{NodeName: "ip-172-20-125-125.ec2.internal"},
				Pods: []*Pod{
					{
						PodRef: struct {
							Name      string "json:\"name\""
							Namespace string "json:\"namespace\""
						}{Name: "aws-xray-daemon-bpmqx", Namespace: "kube-system"},
						EphemeralStorage: &Volume{
							UsedBytes:      36864,
							AvailableBytes: 92321636352,
						},
					},
				},
			},
		},
		{
			Name:      "validate example one test data",
			InputFile: "testdata/example2.yaml",
			WantSummary: &Summary{Node: Node{NodeName: "ip-172-20-96-152.ec2.internal"},
				Pods: []*Pod{
					{
						PodRef: struct {
							Name      string "json:\"name\""
							Namespace string "json:\"namespace\""
						}{Name: "appcache-us-east-1f-6599bdfbcd-lf9j6", Namespace: "rmux"},
						EphemeralStorage: &Volume{
							UsedBytes:      20480,
							AvailableBytes: 92463894528,
						},
					},
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			ex, err := ioutil.ReadFile(tc.InputFile)
			if err != nil {
				t.Fatalf("failed to read test data %+v", err)
			}
			scraper := NewScraper(zap.L(), "", "", 1*time.Microsecond)
			summary, err := scraper.parse(ex)
			if err != nil {
				t.Fatalf("failed to parse test data %+v", err)
			}
			if diff := cmp.Diff(tc.WantSummary, summary); diff != "" {
				t.Errorf("summary mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
