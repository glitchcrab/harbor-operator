package chartmuseum

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/markbates/pkger"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/factories/application"
)

const (
	configName = "config.yaml"
)

var (
	once   sync.Once
	config []byte
)

func InitConfigMaps() {
	file, err := pkger.Open("/assets/templates/chartmuseum/config.yaml")
	if err != nil {
		panic(errors.Wrapf(err, "cannot open ChartMuseum configuration template %s", "/assets/templates/chartmuseum/config.yaml"))
	}
	defer file.Close()

	config, err = ioutil.ReadAll(file)
	if err != nil {
		panic(errors.Wrapf(err, "cannot read ChartMuseum configuration template %s", "/assets/templates/chartmuseum/config.yaml"))
	}
}

// https://github.com/goharbor/harbor/blob/master/make/photon/prepare/templates/chartserver/env.jinja

func (c *ChartMuseum) GetConfigMaps(ctx context.Context) []*corev1.ConfigMap {
	once.Do(InitConfigMaps)

	operatorName := application.GetName(ctx)
	harborName := c.harbor.Name

	return []*corev1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.harbor.NormalizeComponentName(containerregistryv1alpha1.ChartMuseumName),
				Namespace: c.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.ChartMuseumName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			BinaryData: map[string][]byte{
				configName: config,
			},
			Data: map[string]string{
				"PORT":                  fmt.Sprintf("%d", port),
				"STORAGE":               "local",
				"STORAGE_LOCAL_ROOTDIR": "/mnt/chartmuseum",
				"CHART_URL":             fmt.Sprintf("%s/chartrepo", c.harbor.Spec.PublicURL),
			},
		},
	}
}
