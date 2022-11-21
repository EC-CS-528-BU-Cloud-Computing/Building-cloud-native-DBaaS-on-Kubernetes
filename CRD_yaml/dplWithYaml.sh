#!/bin/sh

kubectl apply -f pd.yaml
kubectl apply -f tikv.yaml
kubectl apply -f tidb.yaml