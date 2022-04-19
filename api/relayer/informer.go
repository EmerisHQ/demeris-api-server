package relayer

import "k8s.io/client-go/informers"

type Informer struct {
	Informer  informers.GenericInformer
	Namespace string
}

func NewInformer(i informers.GenericInformer, ns string) *Informer {
	return &Informer{
		Informer:  i,
		Namespace: ns,
	}
}
