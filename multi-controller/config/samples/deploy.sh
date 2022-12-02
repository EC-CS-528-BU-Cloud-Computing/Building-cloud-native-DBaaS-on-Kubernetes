#!/bin/sh

kubectl apply -f tidb-cluster_v1_pd.yaml
kubectl apply -f tidb-cluster_v1_tikv.yaml
kubectl apply -f tidb-cluster_v1_tidb.yaml
kubectl apply -f pd-pv.yaml
kubectl apply -f pd-pvc.yaml
kubectl apply -f tikv-pv.yaml
kubectl apply -f tikv-pvc.yaml