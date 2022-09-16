package consoleplugin

import (
	"fmt"

	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/api/resource"
)

func ConsolePlugin(namespace string) *consolev1alpha1.ConsolePlugin {
	return &consolev1alpha1.ConsolePlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: PluginName,
		},
		Spec: consolev1alpha1.ConsolePluginSpec{
			DisplayName: PluginDisplayName,
			Service: consolev1alpha1.ConsolePluginService{
				Name: PluginName,
				Namespace: namespace,
				Port: int32(PluginPort),
				BasePath: PluginBasePath,
			},
		},
	}
}

func Service(namespace string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: PluginName,
			Namespace: namespace,
			Annotations: map[string]string{
				"service.alpha.openshift.io/serving-cert-secret-name": "plugin-serving-cert",
			},
			Labels: map[string]string{
				"app": PluginName,
    			"app.kubernetes.io/component": PluginName,
    			"app.kubernetes.io/instance": PluginName,
    			"app.kubernetes.io/part-of": PluginName,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name: "9443-tcp",
				Protocol: "TCP",
				Port: PluginPort,
				TargetPort: intstr.IntOrString{IntVal: int32(PluginPort)},
			}},
			Selector: map[string]string {
				"app": PluginName,
			},
			Type: "ClusterIP",
			SessionAffinity: "None",
		},
	}
}

func Deployment(namespace string) *appsv1.Deployment{
	resources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			"cpu": resource.MustParse(CPULimit),
			"memory": resource.MustParse(MemoryLimit),
		},
	}

	volumeMounts := []corev1.VolumeMount{{
		Name: SecretName,
		MountPath: SecretMount,
		ReadOnly: true,
	}, {
		Name: ConfigVolume,
		ReadOnly: true,
		MountPath: ConfigPath,
		SubPath: SubPathConf,
	}}

	volumes := []corev1.Volume{{
		Name: SecretName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: SecretName,
				DefaultMode: getInt32Pointer(420),
			},
		},
	}, {
		Name: ConfigVolume,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: ConfigVolume,
				},
				DefaultMode: getInt32Pointer(420),
			},
		},
	}}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: PluginName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": PluginName,
				"app.kubernetes.io/component": PluginName,
				"app.kubernetes.io/instance": PluginName,
				"app.kubernetes.io/part-of": PluginName,
				"app.openshift.io/runtime-namespace": namespace,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": PluginName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": PluginName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name: PluginName,
						Image: PluginImage,
						Ports: []corev1.ContainerPort{{
							ContainerPort:PluginPort,
							Protocol: "TCP",
						}},
						ImagePullPolicy: "Always",
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: getBoolPointer(false),
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{"ALL"},
							},
						},
						Resources: resources,
						VolumeMounts: volumeMounts,
					}},
					Volumes: volumes,
					RestartPolicy: "Always",
					DNSPolicy: "ClusterFirst",
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: getBoolPointer(true),
						SeccompProfile: &corev1.SeccompProfile{
							Type: "RuntimeDefault",
						},
					},
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{IntVal: int32(25)},
					MaxSurge: &intstr.IntOrString{IntVal: int32(25)},
				},
			},
		},
	}
}

func ConfigMap(namespace string) *corev1.ConfigMap {
	data := map[string]string{
		"nginx.conf": fmt.Sprintf(`
				error_log /dev/stdout info;
				events {}
				http {
					access_log			/dev/stdout;
					include				/etc/nginx/mime.types;
					default_type		application/octet-stream;
					keepalive_timeout	65;
					server {
						listen				%v ssl;
						ssl_certificate		/var/serving-cert/tls.crt;
						ssl_certificate_key	/var/serving-cert/tls.key;
						root				/usr/share/nginx/html;
					}
				}
				`, 9443)}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: ConfigName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": PluginName,
				"app.kubernetes.io/part-of": PluginName,
			},
		},
		Data: data,
	}
}

func getInt32Pointer(value int32) *int32 {
	return &value
}

func getBoolPointer(value bool) *bool {
	return &value
}