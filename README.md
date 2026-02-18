# MediaServer Operator

This MediaServer Operator is an implementation of Kubernetes Operator that allows to launch Flussonic Media Server in k8s with a trivial configuration:

```
apiVersion: media.flussonic.com/v1alpha1
kind: MediaServer
metadata:
  name: watcher
spec:
  image: flussonic/flussonic:v24.02-107
  hostPort: 80
  env:
    - name: LICENSE_KEY
      value: "your-license-key"
```

This Operator will create about 10 different resources and manage them.


## Usage


Как собрать и отладить ветку.

1. В пайплайнах запускаем publish-controller-rolling, это зальет контроллер на докерхаб

```
export KUBECONFIG=`pwd`/kube.yaml
touch kube.yaml
chmod 0600 kube.yaml
kind create cluster
make operator.yaml
# Обрати внимание!!!
# mkdir -p docs/latest docs/26.1.6-2-gadc4be7
# /Users/max/Sites/media-server-operator/bin/kustomize build config/default > docs/26.1.6-2-gadc4be7/operator.yaml
# [ "ml-58393-fix-labels" = "master" ] && cp docs/26.1.6-2-gadc4be7/operator.yaml docs/latest/operator.yaml || true
# Коммитить этот тег не нужно! Он для разработки!
kubectl apply -f docs/26.1.6-2-gadc4be7/operator.yaml
source ./env && kubectl create secret generic flussonic-license \
    --from-literal=license_key="${LICENSE_KEY}" \
    --from-literal=edit_auth="root:password"
kubectl apply -f config/samples/media_v1alpha1_streamer-simple.yaml
```



To test on your laptop with `multipass` VM management tool:

```
./mp-start.sh
```

On your k8s installation:

```
kubectl label nodes streamer flussonic.com/streamer=true
kubectl create secret generic flussonic-license \
    --from-literal=license_key="${LICENSE_KEY}" \
    --from-literal=edit_auth="root:password"


kubectl apply -f https://flussonic.github.io/media-server-operator/latest/operator.yaml
kubectl apply -f config/samples/media_v1alpha1_streamer-simple.yaml
```

Then it will run on your server


## Test

1. Собираем промежуточное обновление: `make deploy`


## Release


1. Мержим ветку в мастер
2. Ставим тег:  `git tag -a v26.2.1 -m 'version 26.2.1'`   (тег формата vYY.M.N)
3. Пушим: `git push --tags`
4. Теперь руками запускаем джобу `publish-controller` на предыдущем перед тегом коммите. Он окажется с аккурат нужным тегом
4. Обновляем ассеты:  `make operator.yaml`

Появится что-то типа:

```
mkdir -p docs/latest docs/26.2.1
/Users/max/Sites/media-server-operator/bin/kustomize build config/default > docs/26.2.1/operator.yaml
[ "master" = "master" ] && cp docs/26.2.1/operator.yaml docs/latest/operator.yaml || true
```

5. Коммитим:  `git add . && git commit -m 'publish 26.2.1' && git push`





# Development of media-server-operator

### Prerequisites
- go version v1.20.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster

**Build and push your image to the location specified by `IMG`:**

```sh
make docker-buildx
make operator.yaml
cp -f operator.yaml docs/
git add docs
```

**NOTE:** This image ought to be published in the public registry.

## Develop locally

```sh
make all install run
```

Another tab:

```sh
kubectl apply -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

