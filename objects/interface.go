package objects

import (
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Object interface {
	// Name is the given name for the object
	Name() string

	//Get return  the client object inside the binded object
	Get() client.Object

	// NewCopy return new deep copy of the object
	NewCopy() Object

	//Containers return list of all available containers inside the object
	Containers() []v1.Container

	// OverrideImage overrides the specific image name which holds the given index
	OverrideImage(containerIndex int, newImage string)
}
