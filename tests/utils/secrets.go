package utils

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	secretPollPeriod   = defaultPollPeriod
	secretPollInterval = defaultPollInterval
)

// NewSecretDefinition provides a function to initialize a Secret data type with the provided options
func NewSecretDefinition(labels, stringData map[string]string, data map[string][]byte, ns, prefix string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefix,
			Namespace:    ns,
			Labels:       labels,
		},
		StringData: stringData,
		Data:       data,
	}
}

// CreateSecretFromDefinition creates and returns a pointer ot a v1.Secret using a provided v1.Secret
func CreateSecretFromDefinition(c *kubernetes.Clientset, secret *v1.Secret) (*v1.Secret, error) {
	err := wait.PollUntilContextTimeout(context.TODO(), secretPollInterval, secretPollPeriod, true, func(ctx context.Context) (done bool, err error) {
		secret, err = c.CoreV1().Secrets(secret.Namespace).Create(ctx, secret, metav1.CreateOptions{})
		// success
		if err == nil {
			return true, nil
		}
		// fail if secret exists.
		if apierrs.IsAlreadyExists(err) {
			return true, err
		}
		// Log non-fatal errors
		klog.Error(errors.Wrapf(err, "Encountered create error for secret \"%s/%s\"", secret.Namespace, secret.Name))
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return secret, nil
}

// DeleteSecret ...
func DeleteSecret(clientSet *kubernetes.Clientset, namespace string, secret v1.Secret) error {
	e := wait.PollUntilContextTimeout(context.TODO(), secretPollInterval, secretPollPeriod, true, func(ctx context.Context) (bool, error) {
		err := clientSet.CoreV1().Secrets(namespace).Delete(ctx, secret.GetName(), metav1.DeleteOptions{})
		if err == nil || apierrs.IsNotFound(err) {
			return true, nil
		}
		return false, nil //keep polling
	})
	return e
}
