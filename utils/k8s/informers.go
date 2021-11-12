package k8s

import (
	appsv1 "github.com/allinbits/starport-operator/api/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
)

func GetInformer(cfg *rest.Config, namespace, resourceType string) (informers.GenericInformer, error) {
	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dc, 0, namespace, nil)
	return factory.ForResource(schema.GroupVersionResource{
		Group:    appsv1.GroupVersion.Group,
		Version:  appsv1.GroupVersion.Version,
		Resource: resourceType,
	}), nil
}

func GetRelayerFromObj(object runtime.Object) (*appsv1.Relayer, error) {
	relayer := &appsv1.Relayer{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		object.(*unstructured.Unstructured).UnstructuredContent(), relayer)
	if err != nil {
		return nil, err
	}

	return relayer, nil
}
