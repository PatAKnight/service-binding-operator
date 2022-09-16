package consoleplugin

const (
	PluginName = "sbo-demo-plugin"
	PluginDisplayName = "OpenShift Console SBO Demo Plugin"
	PluginPort int32 = 9443
	PluginBasePath = "/"
	PluginReplicas int32 = 1
	PluginImage = "quay.io/rh_ee_pknight/sbo-demo-plugin"
	CPULimit = "10m"
	MemoryLimit = "50Mi"
	SecretName = "plugin-serving-cert"
	SecretMount = "/var/serving-cert"
	ConfigVolume = "nginx-conf"
	ConfigPath = "/etc/nginx/nginx.conf"
	SubPathConf = "nginx.conf"
	ConfigName = "nginx-conf"
)