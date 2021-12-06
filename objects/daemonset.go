package objects

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type daemonSet struct {
	Object *appsv1.DaemonSet
}

func (d *daemonSet) Name() string {
	return `DaemonSet`
}

func (d *daemonSet) Get() client.Object {
	return d.Object
}

func (d *daemonSet) NewCopy() Object {
	return &daemonSet{Object: d.Object.DeepCopy()}
}

func (d *daemonSet) Containers() []v1.Container {
	return d.Object.Spec.Template.Spec.Containers
}

func (d *daemonSet) OverrideImage(containerIndex int, newImage string) {
	d.Object.Spec.Template.Spec.Containers[containerIndex].Image = newImage
}

func NewDaemonSetObject() Object {
	return &daemonSet{
		Object: &appsv1.DaemonSet{},
	}
}
