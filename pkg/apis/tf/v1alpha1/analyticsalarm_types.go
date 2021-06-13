package v1alpha1

import (
	"bytes"
	"context"
	"sort"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	configtemplates "github.com/tungstenfabric/tf-operator/pkg/apis/tf/v1alpha1/templates"
	"github.com/tungstenfabric/tf-operator/pkg/certificates"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AnalyticsAlarm is the Schema for the Analytics Alarm API.
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=analyticsalarm,scope=Namespaced
// +kubebuilder:printcolumn:name="Active",type=boolean,JSONPath=`.status.active`
type AnalyticsAlarm struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AnalyticsAlarmSpec   `json:"spec,omitempty"`
	Status AnalyticsAlarmStatus `json:"status,omitempty"`
}

// AnalyticsAlarmList contains a list of AnalyticsAlarm.
// +k8s:openapi-gen=true
type AnalyticsAlarmList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []AnalyticsAlarm `json:"items"`
}

// AnalyticsAlarmSpec is the Spec for the Analytics Alarm API.
// +k8s:openapi-gen=true
type AnalyticsAlarmSpec struct {
	CommonConfiguration  PodConfiguration            `json:"commonConfiguration,omitempty"`
	ServiceConfiguration AnalyticsAlarmConfiguration `json:"serviceConfiguration"`
}

// AnalyticsAlarmConfiguration is the Spec for the Analytics Alarm API.
// +k8s:openapi-gen=true
type AnalyticsAlarmConfiguration struct {
	CassandraInstance              string       `json:"cassandraInstance,omitempty"`
	ZookeeperInstance              string       `json:"zookeeperInstance,omitempty"`
	RabbitmqInstance               string       `json:"rabbitmqInstance,omitempty"`
	ConfigInstance                 string       `json:"configInstance,omitempty"`
	LogFilePath                    string       `json:"logFilePath,omitempty"`
	LogLevel                       string       `json:"logLevel,omitempty"`
	LogLocal                       string       `json:"logLocal,omitempty"`
	AlarmgenRedisAggregateDbOffset *int         `json:"alarmgenRedisAggregateDbOffset,omitempty"`
	AlarmgenPartitions             *int         `json:"alarmgenPartitions,omitempty"`
	AlarmgenIntrospectListenPort   *int         `json:"alarmgenIntrospectListenPort,omitempty"`
	AlarmgenLogFileName            string       `json:"alarmgenLogFileName,omitempty"`
	Containers                     []*Container `json:"containers,omitempty"`
}

// AnalyticsAlarmStatus is the Status for the Analytics Alarm API.
// +k8s:openapi-gen=true
type AnalyticsAlarmStatus struct {
	CommonStatus `json:",inline"`
}

func init() {
	SchemeBuilder.Register(&AnalyticsAlarm{}, &AnalyticsAlarmList{})
}

// CreateConfigMap creates analytics alarm config map
func (c *AnalyticsAlarm) CreateConfigMap(configMapName string,
	client client.Client,
	scheme *runtime.Scheme,
	request reconcile.Request) (*corev1.ConfigMap, error) {

	return CreateConfigMap(configMapName,
		client,
		scheme,
		request,
		"analyticsalarm",
		c)
}

// InstanceConfiguration create config data
func (c *AnalyticsAlarm) InstanceConfiguration(configMapName string,
	podList []corev1.Pod,
	request reconcile.Request,
	client client.Client) error {

	configMapInstanceDynamicConfig := &corev1.ConfigMap{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: configMapName, Namespace: request.Namespace}, configMapInstanceDynamicConfig)
	if err != nil {
		return err
	}

	cassandraNodesInformation, err := NewCassandraClusterConfiguration(c.Spec.ServiceConfiguration.CassandraInstance,
		request.Namespace, client)
	if err != nil {
		return err
	}
	zookeeperNodesInformation, err := NewZookeeperClusterConfiguration(c.Spec.ServiceConfiguration.ZookeeperInstance,
		request.Namespace, client)
	if err != nil {
		return err
	}
	rabbitmqNodesInformation, err := NewRabbitmqClusterConfiguration(c.Spec.ServiceConfiguration.RabbitmqInstance, request.Namespace, client)
	if err != nil {
		return err
	}
	configNodesInformation, err := NewConfigClusterConfiguration(c.Spec.ServiceConfiguration.ConfigInstance, request.Namespace, client)
	if err != nil {
		return err
	}

	var rabbitmqSecretUser string
	var rabbitmqSecretPassword string
	var rabbitmqSecretVhost string
	if rabbitmqNodesInformation.Secret != "" {
		rabbitmqSecret := &corev1.Secret{}
		err = client.Get(context.TODO(), types.NamespacedName{Name: rabbitmqNodesInformation.Secret, Namespace: request.Namespace}, rabbitmqSecret)
		if err != nil {
			return err
		}
		rabbitmqSecretUser = string(rabbitmqSecret.Data["user"])
		rabbitmqSecretPassword = string(rabbitmqSecret.Data["password"])
		rabbitmqSecretVhost = string(rabbitmqSecret.Data["vhost"])
	}

	// Create main common values
	rabbitMqSSLEndpointList := configtemplates.EndpointList(rabbitmqNodesInformation.ServerIPList, rabbitmqNodesInformation.Port)
	sort.Strings(rabbitMqSSLEndpointList)
	rabbitmqSSLEndpointListSpaceSeparated := configtemplates.JoinListWithSeparator(rabbitMqSSLEndpointList, " ")

	configDbEndpointList := configtemplates.EndpointList(cassandraNodesInformation.ServerIPList, cassandraNodesInformation.Port)
	sort.Strings(configDbEndpointList)
	configDbEndpointListSpaceSeparated := configtemplates.JoinListWithSeparator(configDbEndpointList, " ")

	configCollectorEndpointList := configtemplates.EndpointList(configNodesInformation.CollectorServerIPList, configNodesInformation.CollectorPort)
	sort.Strings(configCollectorEndpointList)
	configCollectorEndpointListSpaceSeparated := configtemplates.JoinListWithSeparator(configCollectorEndpointList, " ")

	configApiEndpointList := configtemplates.EndpointList(configNodesInformation.APIServerIPList, configNodesInformation.APIServerPort)
	sort.Strings(configApiEndpointList)
	configApiIPEndpointListSpaceSeparated := configtemplates.JoinListWithSeparator(configNodesInformation.APIServerIPList, " ")

	configApiList := make([]string, len(configNodesInformation.APIServerIPList))
	copy(configApiList, configNodesInformation.APIServerIPList)
	sort.Strings(configApiList)
	configApiIPCommaSeparated := configtemplates.JoinListWithSeparator(configApiList, ",")

	zookeeperEndpointList := configtemplates.EndpointList(zookeeperNodesInformation.ServerIPList, zookeeperNodesInformation.ClientPort)
	sort.Strings(zookeeperEndpointList)
	zookeeperEndpointListSpaceSeparated := configtemplates.JoinListWithSeparator(zookeeperEndpointList, " ")
	zookeeperEndpointListCommaSeparated := configtemplates.JoinListWithSeparator(zookeeperEndpointList, ",")

	zookeeperListLength := len(strings.Split(zookeeperEndpointListSpaceSeparated, " "))
	replicationFactor := 1
	minInsyncReplicas := 1
	if zookeeperListLength == 2 {
		replicationFactor = 2
	} else if zookeeperListLength > 2 {
		replicationFactor = 3
		minInsyncReplicas = 2
	}

	var kafkaServerSpaceSeparatedList string
	var podIPList []string
	for _, pod := range podList {
		podIPList = append(podIPList, pod.Status.PodIP)
	}
	sort.SliceStable(podList, func(i, j int) bool { return podList[i].Status.PodIP < podList[j].Status.PodIP })
	sort.SliceStable(podIPList, func(i, j int) bool { return podIPList[i] < podIPList[j] })

	kafkaServerSpaceSeparatedList = strings.Join(podIPList, ":9092 ") + ":9092"

	kafkaSecret := &corev1.Secret{}
	if err = client.Get(context.TODO(), types.NamespacedName{Name: request.Name + "-secret", Namespace: request.Namespace}, kafkaSecret); err != nil {
		return err
	}

	redisServerSpaceSeparatedList := strings.Join(podIPList, ":6379 ") + ":6379"

	for _, pod := range podList {
		hostname := pod.Annotations["hostname"]
		podIP := pod.Status.PodIP
		instrospectListenAddress := c.Spec.CommonConfiguration.IntrospectionListenAddress(podIP)

		var alarmBuffer bytes.Buffer
		err = configtemplates.AnalyticsAlarmgenConfig.Execute(&alarmBuffer, struct {
			PodIP                          string
			Hostname                       string
			ListenAddress                  string
			InstrospectListenAddress       string
			AlarmgenRedisAggregateDbOffset string
			AlarmgenPartitions             string
			AlarmgenIntrospectListenPort   string
			LogFile                        string
			LogLevel                       string
			LogLocal                       string
			CollectorServers               string
			ZookeeperServers               string
			ConfigServers                  string
			ConfigDbServerList             string
			KafkaServers                   string
			CassandraSslCaCertfile         string
			RabbitmqServerList             string
			RabbitmqVhost                  string
			RabbitmqUser                   string
			RabbitmqPassword               string
			RedisServerList                string
			CAFilePath                     string
		}{
			PodIP:                    podIP,
			Hostname:                 hostname,
			ListenAddress:            podIP,
			InstrospectListenAddress: instrospectListenAddress,
			CollectorServers:         configCollectorEndpointListSpaceSeparated,
			ZookeeperServers:         zookeeperEndpointListSpaceSeparated,
			ConfigServers:            configApiIPEndpointListSpaceSeparated,
			ConfigDbServerList:       configDbEndpointListSpaceSeparated,
			KafkaServers:             kafkaServerSpaceSeparatedList,
			CassandraSslCaCertfile:   certificates.SignerCAFilepath,
			RabbitmqServerList:       rabbitmqSSLEndpointListSpaceSeparated,
			RabbitmqVhost:            rabbitmqSecretVhost,
			RabbitmqUser:             rabbitmqSecretUser,
			RabbitmqPassword:         rabbitmqSecretPassword,
			RedisServerList:          redisServerSpaceSeparatedList,
			CAFilePath:               certificates.SignerCAFilepath,
			// TODO: move to params
			LogLevel: "SYS_DEBUG",
		})
		if err != nil {
			panic(err)
		}
		configMapInstanceDynamicConfig.Data["tf-alarm-gen."+podIP] = alarmBuffer.String()

		myidString := pod.Name[len(pod.Name)-1:]
		myidInt, err := strconv.Atoi(myidString)
		if err != nil {
			return err
		}

		var kafkaBuffer bytes.Buffer
		err = configtemplates.KafkaConfig.Execute(&kafkaBuffer, struct {
			PodIP              string
			BrokerId           string
			Hostname           string
			ZookeeperServers   string
			ReplicationFactor  string
			MinInsyncReplicas  string
			KeystorePassword   string
			TruststorePassword string
			CAFilePath         string
			LogLevel           string
		}{
			PodIP:              podIP,
			BrokerId:           strconv.Itoa(myidInt),
			Hostname:           hostname,
			ZookeeperServers:   zookeeperEndpointListCommaSeparated,
			ReplicationFactor:  strconv.Itoa(replicationFactor),
			MinInsyncReplicas:  strconv.Itoa(minInsyncReplicas),
			KeystorePassword:   string(kafkaSecret.Data["keystorePassword"]),
			TruststorePassword: string(kafkaSecret.Data["truststorePassword"]),
			CAFilePath:         certificates.SignerCAFilepath,
			// TODO: move to params
			LogLevel: "SYS_DEBUG",
		})
		if err != nil {
			panic(err)
		}
		configMapInstanceDynamicConfig.Data["kafka.config."+podIP] = kafkaBuffer.String()

		// TODO: commonize for all services
		var nodemanagerBuffer bytes.Buffer
		err = configtemplates.NodemanagerConfig.Execute(&nodemanagerBuffer, struct {
			Hostname                 string
			PodIP                    string
			ListenAddress            string
			InstrospectListenAddress string
			CollectorServerList      string
			CassandraPort            string
			CassandraJmxPort         string
			CAFilePath               string
			MinimumDiskGB            int
			LogLevel                 string
			LogFile                  string
			LogLocal                 string
		}{
			PodIP:                    podIP,
			Hostname:                 hostname,
			ListenAddress:            podIP,
			InstrospectListenAddress: instrospectListenAddress,
			CassandraPort:            strconv.Itoa(cassandraNodesInformation.CQLPort),
			CassandraJmxPort:         strconv.Itoa(cassandraNodesInformation.JMXPort),
			CAFilePath:               certificates.SignerCAFilepath,
			CollectorServerList:      configCollectorEndpointListSpaceSeparated,
			// TODO: move to params
			LogLevel: "SYS_DEBUG",
		})
		if err != nil {
			panic(err)
		}
		configMapInstanceDynamicConfig.Data["analytics-alarm-nodemgr.conf."+podIP] = nodemanagerBuffer.String()
		// empty env as no db tracking
		configMapInstanceDynamicConfig.Data["analytics-alarm-nodemgr.env."+podIP] = ""

		// TODO: commonize for all services
		var vnciniBuffer bytes.Buffer
		err = configtemplates.ConfigAPIVNC.Execute(&vnciniBuffer, struct {
			APIServerList          string
			APIServerPort          string
			CAFilePath             string
			AuthMode               AuthenticationMode
			KeystoneAuthParameters *KeystoneAuthParameters
			PodIP                  string
		}{
			APIServerList:          configApiIPCommaSeparated,
			APIServerPort:          strconv.Itoa(configNodesInformation.APIServerPort),
			CAFilePath:             certificates.SignerCAFilepath,
			AuthMode:               c.Spec.CommonConfiguration.AuthParameters.AuthMode,
			KeystoneAuthParameters: c.Spec.CommonConfiguration.AuthParameters.KeystoneAuthParameters,
			PodIP:                  podIP,
		})
		if err != nil {
			panic(err)
		}
		configMapInstanceDynamicConfig.Data["vnc_api_lib.ini."+podIP] = vnciniBuffer.String()
	}

	// update with provisioner configs
	UpdateProvisionerConfigMapData("analyticsalarm-provisioner", configApiIPCommaSeparated,
		c.Spec.CommonConfiguration.AuthParameters, configMapInstanceDynamicConfig)

	return client.Update(context.TODO(), configMapInstanceDynamicConfig)
}

// CreateSecret creates a secret.
func (c *AnalyticsAlarm) CreateSecret(secretName string,
	client client.Client,
	scheme *runtime.Scheme,
	request reconcile.Request) (*corev1.Secret, error) {
	return CreateSecret(secretName,
		client,
		scheme,
		request,
		"kafka",
		c)
}

//PodsCertSubjects gets list of Vrouter pods certificate subjets which can be passed to the certificate API
func (c *AnalyticsAlarm) PodsCertSubjects(domain string, podList []corev1.Pod) []certificates.CertificateSubject {
	var altIPs PodAlternativeIPs
	return PodsCertSubjects(domain, podList, c.Spec.CommonConfiguration.HostNetwork, altIPs)
}

// PodIPListAndIPMapFromInstance gets a list with POD IPs and a map of POD names and IPs.
func (c *AnalyticsAlarm) PodIPListAndIPMapFromInstance(instanceType string, request reconcile.Request, reconcileClient client.Client) ([]corev1.Pod, map[string]string, error) {
	return PodIPListAndIPMapFromInstance(instanceType, request, reconcileClient, "")
}

// SetInstanceActive sets instance to active.
func (c *AnalyticsAlarm) SetInstanceActive(client client.Client, activeStatus *bool, degradedStatus *bool, sts *appsv1.StatefulSet, request reconcile.Request) error {
	return SetInstanceActive(client, activeStatus, degradedStatus, sts, request, c)
}

// CommonStartupScript prepare common run service script
//  command - is a final command to run
//  configs - config files to be waited for and to be linked from configmap mount
//   to a destination config folder (if destination is empty no link be done, only wait), e.g.
//   { "api.${POD_IP}": "", "vnc_api.ini.${POD_IP}": "vnc_api.ini"}
func (c *AnalyticsAlarm) CommonStartupScript(command string, configs map[string]string) string {
	return CommonStartupScript(command, configs)
}
