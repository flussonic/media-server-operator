#!/bin/bash

set -ex

if [ -f env ]; then
    set -a
    source ./env
    set +a
fi

if [ -z "$LICENSE_KEY" ]; then
    read -p "Enter Flussonic license key: "  LICENSE_KEY
fi


multipass launch --name streamer --cpus 1 --memory 1024M --disk 5G lts
multipass exec streamer -- sudo /bin/sh -c 'curl -sfL https://get.k3s.io | sh -'
multipass exec streamer -- sudo mkdir -p /storage

token=$(multipass exec streamer sudo cat /var/lib/rancher/k3s/server/node-token)
plane_ip=$(multipass info streamer | grep -i ip | awk '{print $2}')
rm -f k3s.yaml
multipass exec streamer sudo cat /etc/rancher/k3s/k3s.yaml |sed "s/127.0.0.1/${plane_ip}/" > k3s.yaml
chmod 0400 k3s.yaml
export KUBECONFIG=`pwd`/k3s.yaml

kubectl label nodes streamer flussonic.com/streamer=true
kubectl create secret generic flussonic-license \
    --from-literal=license_key="${LICENSE_KEY}" \
    --from-literal=edit_auth="root:password"  # root:password



# kubectl apply -f https://flussonic.github.io/media-server-operator/operator.yaml
# kubectl apply -f config/samples/media_v1alpha1_mediaserver.yaml


echo "Streamer ready: http://${plane_ip}"
