# bin/sh

sudo kind delete cluster
sudo kind create cluster

sudo make install
sudo sh ./config/samples/deploy.sh 
sudo make run

# new window
sudo kubectl get pods
