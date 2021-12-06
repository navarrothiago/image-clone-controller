package objects

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type deployment struct {
	Object *appsv1.Deployment
}

func (d *deployment) Name() string {
	return `Deployment`
}

func (d *deployment) Get() client.Object {
	return d.Object
}

func (d *deployment) NewCopy() Object {
	return &deployment{Object: d.Object.DeepCopy()}
}

func (d *deployment) Containers() []v1.Container {
	return d.Object.Spec.Template.Spec.Containers
}

func (d *deployment) OverrideImage(containerIndex int, newImage string) {
	d.Object.Spec.Template.Spec.Containers[containerIndex].Image = newImage
}

func NewDeploymentObject() Object {
	return &deployment{
		Object: &appsv1.Deployment{},
	}
}
