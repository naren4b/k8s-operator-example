# k8s-operator-example
Write a Kubernetes CRD- `ImageArray` that will have array of full image path 
Write an Kubernetes Operator which watch for the ImageArray CRs 
If any ImageArray CR deployed then it will create a kubernetes job  which will pull these images tag it and push to the local registry 
INPUT: 
SOURCE=registry-1
TARGET=registry-2

docker login $SOURCE read from the secret source-registry-image-pull-secret
docker login $TARGET read from the secret target-registry-image-push-secret

OUT put print each image pushed successfully or not 





