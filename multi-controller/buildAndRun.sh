# bin/sh

sudo kind delete cluster
sudo kind create cluster

sudo make install
cd config/samples
sudo sh deploy.sh 
sudo sh prometheusDepl.sh 
cd ../..
sudo make run

# new window
# sudo kubectl get pods
