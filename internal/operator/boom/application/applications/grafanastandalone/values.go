package grafanastandalone

type Rbac struct {
	Create                bool          `yaml:"create"`
	PspEnabled            bool          `yaml:"pspEnabled"`
	PspUseAppArmor        bool          `yaml:"pspUseAppArmor"`
	Namespaced            bool          `yaml:"namespaced"`
	ExtraRoleRules        []interface{} `yaml:"extraRoleRules"`
	ExtraClusterRoleRules []interface{} `yaml:"extraClusterRoleRules"`
}
type ServiceAccount struct {
	Create   bool        `yaml:"create"`
	Name     interface{} `yaml:"name"`
	NameTest interface{} `yaml:"nameTest"`
}
type DeploymentStrategy struct {
	Type string `yaml:"type"`
}
type HTTPGet struct {
	Path string `yaml:"path"`
	Port int    `yaml:"port"`
}
type ReadinessProbe struct {
	HTTPGet *HTTPGet `yaml:"httpGet"`
}
type LivenessProbe struct {
	HTTPGet             *HTTPGet `yaml:"httpGet"`
	InitialDelaySeconds int      `yaml:"initialDelaySeconds"`
	TimeoutSeconds      int      `yaml:"timeoutSeconds"`
	FailureThreshold    int      `yaml:"failureThreshold"`
}
type Image struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
	PullPolicy string `yaml:"pullPolicy"`
}
type TestFramework struct {
	Enabled         bool   `yaml:"enabled"`
	Image           string `yaml:"image"`
	Tag             string `yaml:"tag"`
	SecurityContext struct {
	} `yaml:"securityContext"`
}
type SecurityContext struct {
	RunAsUser int `yaml:"runAsUser"`
	FsGroup   int `yaml:"fsGroup"`
}
type DownloadDashboardsImage struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
	PullPolicy string `yaml:"pullPolicy"`
}
type DownloadDashboards struct {
	Env *Env `yaml:"env"`
}
type Env struct {
}
type Service struct {
	Type        string            `yaml:"type"`
	Port        int               `yaml:"port"`
	TargetPort  int               `yaml:"targetPort"`
	Annotations map[string]string `yaml:"annotations"`
	Labels      map[string]string `yaml:"labels"`
	PortName    string            `yaml:"portName"`
}
type Ingress struct {
	Enabled     bool              `yaml:"enabled"`
	Annotations map[string]string `yaml:"annotations"`
	Labels      map[string]string `yaml:"labels"`
	Path        string            `yaml:"path"`
	Hosts       []string          `yaml:"hosts"`
	ExtraPaths  []interface{}     `yaml:"extraPaths"`
	TLS         []interface{}     `yaml:"tls"`
}
type Persistence struct {
	Type        string   `yaml:"type"`
	Enabled     bool     `yaml:"enabled"`
	AccessModes []string `yaml:"accessModes"`
	Size        string   `yaml:"size"`
	Finalizers  []string `yaml:"finalizers"`
}
type InitChownData struct {
	Enabled   bool     `yaml:"enabled"`
	Image     *Image   `yaml:"image"`
	Resources struct{} `yaml:"resources"`
}
type Admin struct {
	ExistingSecret string `yaml:"existingSecret"`
	UserKey        string `yaml:"userKey"`
	PasswordKey    string `yaml:"passwordKey"`
}
type Datasources struct {
	Datasources *Datasourcesyaml `yaml:"datasources.yaml"`
}
type Datasourcesyaml struct {
	APIVersion  int64         `yaml:"apiVersion"`
	Datasources []*Datasource `yaml:"datasources"`
}
type Datasource struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	URL       string `yaml:"url"`
	Access    string `yaml:"access"`
	IsDefault bool   `yaml:"isDefault"`
}
type Dashboards struct {
	Dashboards map[string]map[string]*DashboardFile `yaml:"dashboards"`
}
type DashboardFile struct {
	File string `yaml:"file"`
}
type Paths struct {
	Data         string `yaml:"data"`
	Logs         string `yaml:"logs"`
	Plugins      string `yaml:"plugins"`
	Provisioning string `yaml:"provisioning"`
}
type Analytics struct {
	CheckForUpdates bool `yaml:"check_for_updates"`
}
type Log struct {
	Mode string `yaml:"mode"`
}
type GrafanaNet struct {
	URL string `yaml:"url"`
}
type GrafanaIni struct {
	Paths      *Paths      `yaml:"paths"`
	Analytics  *Analytics  `yaml:"analytics"`
	Log        *Log        `yaml:"log"`
	GrafanaNet *GrafanaNet `yaml:"grafana_net"`
}
type Ldap struct {
	Enabled        bool   `yaml:"enabled"`
	ExistingSecret string `yaml:"existingSecret"`
	Config         string `yaml:"config"`
}
type SMTP struct {
	ExistingSecret string `yaml:"existingSecret"`
	UserKey        string `yaml:"userKey"`
	PasswordKey    string `yaml:"passwordKey"`
}
type ProviderSidecar struct {
	Name          string `yaml:"name"`
	Orgid         int    `yaml:"orgid"`
	Folder        string `yaml:"folder"`
	Type          string `yaml:"type"`
	DisableDelete bool   `yaml:"disableDelete"`
}
type DashboardsSidecar struct {
	Enabled           bool             `yaml:"enabled"`
	Label             string           `yaml:"label"`
	Folder            string           `yaml:"folder"`
	DefaultFolderName interface{}      `yaml:"defaultFolderName"`
	SearchNamespace   interface{}      `yaml:"searchNamespace"`
	Provider          *ProviderSidecar `yaml:"provider"`
}
type DatasourcesSidecar struct {
	Enabled         bool        `yaml:"enabled"`
	Label           string      `yaml:"label"`
	SearchNamespace interface{} `yaml:"searchNamespace"`
}
type Sidecar struct {
	Image           string              `yaml:"image"`
	ImagePullPolicy string              `yaml:"imagePullPolicy"`
	Resources       struct{}            `yaml:"resources"`
	Dashboards      *DashboardsSidecar  `yaml:"dashboards"`
	Datasources     *DatasourcesSidecar `yaml:"datasources"`
}
type DashboardProviders struct {
	Providers *Providersyaml `yaml:"dashboardproviders.yaml"`
}
type Providersyaml struct {
	APIVersion int64       `yaml:"apiVersion"`
	Providers  []*Provider `yaml:"providers"`
}
type Provider struct {
	Name            string            `yaml:"name"`
	OrgID           int               `yaml:"ordId"`
	Folder          string            `yaml:"folder,omitempty"`
	Type            string            `yaml:"type"`
	DisableDeletion bool              `yaml:"disableDeletion"`
	Editable        bool              `yaml:"editable"`
	Options         map[string]string `yaml:"options"`
}

type Values struct {
	FullnameOverride        string                   `yaml:"fullnameOverride,omitempty"`
	Rbac                    *Rbac                    `yaml:"rbac"`
	ServiceAccount          *ServiceAccount          `yaml:"serviceAccount"`
	Replicas                int                      `yaml:"replicas"`
	PodDisruptionBudget     struct{}                 `yaml:"podDisruptionBudget"`
	DeploymentStrategy      *DeploymentStrategy      `yaml:"deploymentStrategy"`
	ReadinessProbe          *ReadinessProbe          `yaml:"readinessProbe"`
	LivenessProbe           *LivenessProbe           `yaml:"livenessProbe"`
	Image                   *Image                   `yaml:"image"`
	TestFramework           *TestFramework           `yaml:"testFramework"`
	SecurityContext         *SecurityContext         `yaml:"securityContext"`
	ExtraConfigmapMounts    []interface{}            `yaml:"extraConfigmapMounts"`
	ExtraEmptyDirMounts     []interface{}            `yaml:"extraEmptyDirMounts"`
	DownloadDashboardsImage *DownloadDashboardsImage `yaml:"downloadDashboardsImage"`
	DownloadDashboards      *DownloadDashboards      `yaml:"downloadDashboards"`
	PodPortName             string                   `yaml:"podPortName"`
	Service                 *Service                 `yaml:"service"`
	Ingress                 *Ingress                 `yaml:"ingress"`
	Resources               struct{}                 `yaml:"resources"`
	NodeSelector            map[string]string        `yaml:"nodeSelector"`
	Tolerations             []interface{}            `yaml:"tolerations"`
	Affinity                map[string]string        `yaml:"affinity"`
	ExtraInitContainers     []interface{}            `yaml:"extraInitContainers"`
	ExtraContainers         string                   `yaml:"extraContainers"`
	Persistence             *Persistence             `yaml:"persistence"`
	InitChownData           *InitChownData           `yaml:"initChownData"`
	AdminUser               string                   `yaml:"adminUser"`
	AdminPassword           string                   `yaml:"adminPassword"`
	Admin                   *Admin                   `yaml:"admin"`
	Env                     map[string]string        `yaml:"env"`
	EnvFromSecret           string                   `yaml:"envFromSecret"`
	EnvRenderSecret         struct{}                 `yaml:"envRenderSecret"`
	ExtraSecretMounts       []interface{}            `yaml:"extraSecretMounts"`
	ExtraVolumeMounts       []interface{}            `yaml:"extraVolumeMounts"`
	Plugins                 []interface{}            `yaml:"plugins"`
	Datasources             *Datasources             `yaml:"datasources"`
	Notifiers               struct{}                 `yaml:"notifiers"`
	DashboardProviders      *DashboardProviders      `yaml:"dashboardProviders"`
	Dashboards              *Dashboards              `yaml:"dashboards"`
	DashboardsConfigMaps    map[string]string        `yaml:"dashboardsConfigMaps"`
	GrafanaIni              *GrafanaIni              `yaml:"grafana.ini"`
	Ldap                    *Ldap                    `yaml:"ldap"`
	SMTP                    *SMTP                    `yaml:"smtp"`
	Sidecar                 *Sidecar                 `yaml:"sidecar"`
}
