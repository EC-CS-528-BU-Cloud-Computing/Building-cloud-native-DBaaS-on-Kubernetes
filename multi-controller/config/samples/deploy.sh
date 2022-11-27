#!/bin/sh

kubectl apply -f tidb-cluster_v1_pd.yaml
kubectl apply -f tidb-cluster_v1_tikv.yaml
kubectl apply -f tidb-cluster_v1_tidb.yaml