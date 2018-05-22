package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/imdario/mergo"
	"gopkg.in/couchbase/gocb.v1"
	log "gopkg.in/inconshreveable/log15.v2"
)

type dataStore struct {
	bucket *gocb.Bucket
}

func newDataStore() *dataStore {
	hostname := os.Getenv("CB_HOST")
	if hostname == "" {
		log.Error("missing Couchbase Server hostname")
		os.Exit(1)
	}
	password := os.Getenv("CB_PASS")
	if password == "" {
		log.Error("missing password")
		os.Exit(1)
	}

	connSpecStr := fmt.Sprintf("couchbase://%s", hostname)
	cluster, err := gocb.Connect(connSpecStr)
	if err != nil {
		log.Error("failed to connect to Couchbase Server", "err", err)
		os.Exit(1)
	}

	bucket, err := cluster.OpenBucket("weekly", password)
	if err != nil {
		log.Error("failed to connect to bucket", "err", err)
		os.Exit(1)
	}

	return &dataStore{bucket}
}

type Build struct {
	Build string `json:"build"`
}

func (d *dataStore) getBuilds() (*[]string, error) {
	var builds []string

	query := gocb.NewN1qlQuery(
		"SELECT DISTINCT `build` " +
			"FROM weekly " +
			"WHERE `build` IS NOT MISSING " +
			"ORDER BY `build` DESC;")

	rows, err := ds.bucket.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}

	var row Build
	for rows.Next(&row) {
		builds = append(builds, row.Build)
	}
	return &builds, nil
}

// TestStatus characterizes the status of test execution in Jenkins
type TestStatus struct {
	Failed  int `json:"failed"`
	Passed  int `json:"passed"`
	Missing int `json:"missing"`
	Total   int `json:"total"`
}

// JiraStatus captures a snapshot of JIRA tickets
type JiraStatus struct {
	Created  int `json:"created"`
	Open     int `json:"open"`
	Resolved int `json:"resolved"`
}

// MetricStatus reports anomalies in collected metrics
type MetricStatus struct {
	Changed   int `json:"changed"`
	Collected int `json:"collected"`
}

// KpiStatus records
type KpiStatus struct {
	Passed   int `json:"passed"`
	Violated int `json:"violated"`
	Defined  int `json:"defined"`
}

// Status aggregates all build indicators for a given build
type Status struct {
	Build        string       `json:"build"`
	Component    string       `json:"component"`
	JiraStatus   JiraStatus   `json:"jira_status"`
	KpiStatus    KpiStatus    `json:"kpi_status"`
	TestStatus   TestStatus   `json:"test_status"`
	MetricStatus MetricStatus `json:"metric_status"`
}

func hash(strings ...string) string {
	h := md5.New()
	for _, s := range strings {
		h.Write([]byte(s))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (d *dataStore) getStatus(docId string) (Status, error) {
	var status Status
	_, err := d.bucket.Get(docId, &status)
	return status, err
}

func (d *dataStore) updateStatus(status Status) error {
	docId := hash(status.Component, status.Build)

	currStatus, err := d.getStatus(docId)
	if err != gocb.ErrKeyNotFound {
		mergo.Merge(&status, currStatus)
	}

	_, err = d.bucket.Upsert(docId, status, 0)
	if err != nil {
		log.Error("failed to update status", "err", err)
	}
	return err
}

func (d *dataStore) getBuildStatus(build string) (*[]Status, error) {
	var statuses []Status

	query := gocb.NewN1qlQuery(
		"SELECT RAW weekly " +
			"FROM weekly " +
			"WHERE `build` = $1 " +
			"ORDER BY `component`;")

	params := []interface{}{build}

	rows, err := ds.bucket.ExecuteN1qlQuery(query, params)
	if err != nil {
		return nil, err
	}

	var status Status
	for rows.Next(&status) {
		status.TestStatus.Missing = status.TestStatus.Total -
			status.TestStatus.Passed - status.TestStatus.Failed
		statuses = append(statuses, status)
	}

	return &statuses, nil
}
