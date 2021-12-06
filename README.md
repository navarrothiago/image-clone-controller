# Image Clone Controller

Image Clone Controller is a kubernetes controller which watches the applications and “caches” the images by re-uploading to our own registry repository and reconfiguring the applications to use these copies. The controller only watches events from Deployment and DaemonSet kinds. 

## Problem 

Kubernetes cluster can run applications. These applications will often use publicly available container images, like official images of popular programs, e.g. Jenkins, PostgreSQL, and so on. Since the images reside in repositories over which we have no control, it is possible that the owner of the repo deletes the image while our pods are configured to use it. In the case of a subsequent node rotation, the locally cached copies of the images would be deleted and Kubernetes would be unable to re-download them in order to re-provision the applications.

## Goal

Safe against the risk of public container images disappearing from the registry while we use them, breaking the deployments.

## How to Use

The implementation is leverated by the [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder). 

### Build Image

```
make docker-build IMG=navarrothiago/image-clone-controller
```

> Feel free to change the username `navarrothiago` and docker container image name `image-clone-controller`.

### Push Image

In order to push the built image to the repository, run:
```
make docker-push IMG=navarrothiago/image-clone-controller
```

### Deployment

Make sure your cluster is already running. If you want to test locally, you can use minikube.

```bash
# Download Minikube
curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && chmod +x minikube

# Install on your machine
sudo mkdir -p /usr/local/bin/
sudo install minikube /usr/local/bin/

# Start Minikube cluster
minikube start
```

Deploy the image on your cluster:

```
make deploy IMG=navarrothiago/image-clone-controller
```

The controller depends on the docker server, username and password in order to authenticate in the registry API and push the images to the "cache" registry. Change the following command according to your configuration:

```
kubectl create secret --namespace=image-clone-controller-system generic docker-registry-credentials
  --from-literal=docker-server=index.docker.io \
  --from-literal=docker-username=navarrothiago \
  --from-literal=docker-password=*******
```

### Clean-up

Remove the `image-clone-controller` from your cluster:
```
make undeploy
```

## Development

You can execute the controller without deploying the docker image into your cluster. This step is useful during the development phase, because you don't need to build and deploy the docker image. Before execute the command below, make sure you have already configured your kubernetes cluster.

```
make run DOCKER_REGISTRY=index.docker.io DOCKER_USERNAME=navarrothiago DOCKER_PASSWORD=*******
```

>  Do not forget to change the `DOCKER_REGISTRY`, `DOCKER_USERNAME` and `DOCKER_PASSWORD`.

## References

- https://github.com/kubernetes-sigs/controller-runtime
- https://github.com/kubernetes-sigs/controller-runtime/tree/master/examples/builtins
- https://github.com/google/go-containerregistry/blob/master/pkg/v1/remote/README.md
- https://book.kubebuilder.io/
- https://godoc.org/github.com/google/go-containerregistry/pkg/v1/remote
- https://master.sdk.operatorframework.io/docs/building-operators/golang/references/event-filtering/
- https://github.com/wgarunap/kube-image-clone-controller
- https://github.com/dev4devs-com/memcached-operator
- https://dev4devs.com/2020/08/16/how-to-getting-started-develop-go-operators-from-scratch-with-sdk-1-0/
