#!/bin/sh

kubectl create -f https://raw.githubusercontent.com/pingcap/tidb-operator/master/manifests/crd.yaml

helm repo add pingcap https://charts.pingcap.org/

kubectl create namespace tidb-admin

helm install --namespace tidb-admin tidb-operator pingcap/tidb-operator --version v1.4.0-beta.1

kubectl get pods --namespace tidb-admin -l app.kubernetes.io/instance=tidb-operator

kubectl create namespace tidb-cluster && \
    kubectl -n tidb-cluster apply -f https://raw.githubusercontent.com/pingcap/tidb-operator/master/examples/basic/tidb-cluster.yaml

kubectl -n tidb-cluster apply -f https://raw.githubusercontent.com/pingcap/tidb-operator/master/examples/basic/tidb-monitor.yaml