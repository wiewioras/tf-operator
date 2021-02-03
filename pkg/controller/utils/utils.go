package utils

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/Juniper/contrail-operator/pkg/apis/contrail/v1alpha1"
)

var log = logf.Log.WithName("utils")
var reqLogger = log.WithValues()

// const defines the Group constants.
const (
	CASSANDRA   = "Cassandra.contrail.juniper.net"
	ZOOKEEPER   = "Zookeeper.contrail.juniper.net"
	RABBITMQ    = "Rabbitmq.contrail.juniper.net"
	CONFIG      = "Config.contrail.juniper.net"
	CONTROL     = "Control.contrail.juniper.net"
	WEBUI       = "Webui.contrail.juniper.net"
	VROUTER     = "Vrouter.contrail.juniper.net"
	KUBEMANAGER = "Kubemanager.contrail.juniper.net"
	MANAGER     = "Manager.contrail.juniper.net"
	REPLICASET  = "ReplicaSet.apps"
	DEPLOYMENT  = "Deployment.apps"
)

func RemoveIndex(s []corev1.Container, index int) []corev1.Container {
	return append(s[:index], s[index+1:]...)
}

// WebuiGroupKind returns group kind.
func WebuiGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(WEBUI)
}

// VrouterGroupKind returns group kind.
func VrouterGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(VROUTER)
}

// ControlGroupKind returns group kind.
func ControlGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(CONTROL)
}

// ConfigGroupKind returns group kind.
func ConfigGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(CONFIG)
}

// KubemanagerGroupKind returns group kind.
func KubemanagerGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(KUBEMANAGER)
}

// CassandraGroupKind returns group kind.
func CassandraGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(CASSANDRA)
}

// ZookeeperGroupKind returns group kind.
func ZookeeperGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(ZOOKEEPER)
}

// RabbitmqGroupKind returns group kind.
func RabbitmqGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(RABBITMQ)
}

// ReplicaSetGroupKind returns group kind.
func ReplicaSetGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(REPLICASET)
}

// ManagerGroupKind returns group kind.
func ManagerGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(MANAGER)
}

// DeploymentGroupKind returns group kind.
func DeploymentGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(DEPLOYMENT)
}

// DeploymentStatusChange monitors per application size change.
func DeploymentStatusChange(appGroupKind schema.GroupKind) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldDeployment, ok := e.ObjectOld.(*appsv1.Deployment)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			newDeployment, ok := e.ObjectNew.(*appsv1.Deployment)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			isOwner := false
			for _, owner := range newDeployment.ObjectMeta.OwnerReferences {
				if *owner.Controller {
					groupVersionKind := schema.FromAPIVersionAndKind(owner.APIVersion, owner.Kind)
					if appGroupKind == groupVersionKind.GroupKind() {
						isOwner = true
					}
				}
			}
			if (oldDeployment.Status.ReadyReplicas != newDeployment.Status.ReadyReplicas) && isOwner {
				return true
			}
			return false
		},
	}
}

// STSStatusChange monitors per application size change.
func STSStatusChange(appGroupKind schema.GroupKind) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldSTS, ok := e.ObjectOld.(*appsv1.StatefulSet)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			newSTS, ok := e.ObjectNew.(*appsv1.StatefulSet)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			isOwner := false
			for _, owner := range newSTS.ObjectMeta.OwnerReferences {
				if *owner.Controller {
					groupVersionKind := schema.FromAPIVersionAndKind(owner.APIVersion, owner.Kind)
					if appGroupKind == groupVersionKind.GroupKind() {
						isOwner = true
					}
				}
			}
			if (oldSTS.Status.ReadyReplicas != newSTS.Status.ReadyReplicas) && isOwner {
				return true
			}
			return false
		},
	}
}

// DSStatusChange monitors per application size change.
func DSStatusChange(appGroupKind schema.GroupKind) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldDS, ok := e.ObjectOld.(*appsv1.DaemonSet)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			newDS, ok := e.ObjectNew.(*appsv1.DaemonSet)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			isOwner := false
			for _, owner := range newDS.ObjectMeta.OwnerReferences {
				if *owner.Controller {
					groupVersionKind := schema.FromAPIVersionAndKind(owner.APIVersion, owner.Kind)
					if appGroupKind == groupVersionKind.GroupKind() {
						isOwner = true
					}
				}
			}
			if (oldDS.Status.NumberReady != newDS.Status.NumberReady) && isOwner {
				return true
			}
			return false
		},
	}
}

// PodStatusChange monitors per application size change.
func PodStatusChange(appGroupKind schema.GroupKind) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldPod, ok := e.ObjectOld.(*corev1.Pod)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			newPod, ok := e.ObjectNew.(*corev1.Pod)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			isOwner := false
			oldReady := true
			newReady := true
			for _, owner := range newPod.ObjectMeta.OwnerReferences {
				if *owner.Controller {
					groupVersionKind := schema.FromAPIVersionAndKind(owner.APIVersion, owner.Kind)
					if appGroupKind == groupVersionKind.GroupKind() {
						isOwner = true
					}
				}
			}
			for _, containerStatus := range oldPod.Status.ContainerStatuses {
				if !containerStatus.Ready {
					oldReady = false
					break
				}

			}
			for _, containerStatus := range newPod.Status.ContainerStatuses {
				if !containerStatus.Ready {
					newReady = false
					break
				}
			}
			if (oldReady != newReady) && isOwner {
				return true
			}
			return false
		},
	}
}

// PodIPChange returns predicate function based on group kind.
func PodIPChange(appLabel map[string]string) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			for key, value := range e.MetaOld.GetLabels() {
				if appLabel[key] == value {
					oldPod, ok := e.ObjectOld.(*corev1.Pod)
					if !ok {
						reqLogger.Info("type conversion mismatch")
					}
					newPod, ok := e.ObjectNew.(*corev1.Pod)
					if !ok {
						reqLogger.Info("type conversion mismatch")
					}
					return oldPod.Status.PodIP != newPod.Status.PodIP
				}
			}
			return false
		},
	}
}

// PodInitStatusChange returns predicate function based on group kind.
func PodInitStatusChange(appLabel map[string]string) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			for key, value := range e.MetaOld.GetLabels() {
				if appLabel[key] == value {
					oldPod, ok := e.ObjectOld.(*corev1.Pod)
					if !ok {
						reqLogger.Info("type conversion mismatch")
					}
					newPod, ok := e.ObjectNew.(*corev1.Pod)
					if !ok {
						reqLogger.Info("type conversion mismatch")
					}
					newPodReady := true
					oldPodReady := true
					if newPod.Status.InitContainerStatuses == nil {
						newPodReady = false
					}
					if oldPod.Status.InitContainerStatuses == nil {
						oldPodReady = false
					}
					for _, initContainerStatus := range newPod.Status.InitContainerStatuses {
						if initContainerStatus.Name == "init" {
							if !initContainerStatus.Ready {
								newPodReady = false
							}
						}
					}
					for _, initContainerStatus := range oldPod.Status.InitContainerStatuses {
						if initContainerStatus.Name == "init" {
							if !initContainerStatus.Ready {
								oldPodReady = false
							}
						}
					}
					return newPodReady != oldPodReady
				}
			}
			return false
		},
	}
}

// PodInitRunning returns predicate function based on group kind.
func PodInitRunning(appLabel map[string]string) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			for key, value := range e.MetaOld.GetLabels() {
				if appLabel[key] == value {
					oldPod, ok := e.ObjectOld.(*corev1.Pod)
					if !ok {
						reqLogger.Info("type conversion mismatch")
					}
					newPod, ok := e.ObjectNew.(*corev1.Pod)
					if !ok {
						reqLogger.Info("type conversion mismatch")
					}
					newPodRunning := true
					oldPodRunning := true
					if newPod.Status.InitContainerStatuses == nil {
						newPodRunning = false
					}
					if oldPod.Status.InitContainerStatuses == nil {
						oldPodRunning = false
					}
					for _, initContainerStatus := range newPod.Status.InitContainerStatuses {
						if initContainerStatus.Name == "init" {
							if initContainerStatus.State.Running == nil {
								newPodRunning = false
							}
						}
					}
					for _, initContainerStatus := range oldPod.Status.InitContainerStatuses {
						if initContainerStatus.Name == "init" {
							if initContainerStatus.State.Running == nil {
								oldPodRunning = false
							}
						}
					}
					return newPodRunning != oldPodRunning
				}
			}
			return false
		},
	}
}

// CassandraActiveChange returns predicate function based on group kind.
func CassandraActiveChange() predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldCassandra, ok := e.ObjectOld.(*v1alpha1.Cassandra)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			newCassandra, ok := e.ObjectNew.(*v1alpha1.Cassandra)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			return !oldCassandra.Status.Active && newCassandra.Status.Active
		},
	}
}

// ConfigActiveChange returns predicate function based on group kind.
func ConfigActiveChange() predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldConfig, ok := e.ObjectOld.(*v1alpha1.Config)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			newConfig, ok := e.ObjectNew.(*v1alpha1.Config)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			newConfigActive := false
			oldConfigActive := false
			if newConfig.Status.Active != nil {
				newConfigActive = *newConfig.Status.Active
			}
			if oldConfig.Status.Active != nil {
				oldConfigActive = *oldConfig.Status.Active
			}
			if !oldConfigActive && newConfigActive {
				return true
			}
			return false

		},
	}
}

// VrouterActiveChange returns predicate function based on group kind.
func VrouterActiveChange() predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldVrouter, ok := e.ObjectOld.(*v1alpha1.Vrouter)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			newVrouter, ok := e.ObjectNew.(*v1alpha1.Vrouter)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			newVrouterActive := false
			oldVrouterActive := false
			if newVrouter.Status.Active != nil {
				newVrouterActive = *newVrouter.Status.Active
			}
			if oldVrouter.Status.Active != nil {
				oldVrouterActive = *oldVrouter.Status.Active
			}
			if !oldVrouterActive && newVrouterActive {
				return true
			}
			return false

		},
	}
}

// ControlActiveChange returns predicate function based on group kind.
func ControlActiveChange() predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldConfig, ok := e.ObjectOld.(*v1alpha1.Control)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			newConfig, ok := e.ObjectNew.(*v1alpha1.Control)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			newConfigActive := false
			oldConfigActive := false
			if newConfig.Status.Active != nil {
				newConfigActive = *newConfig.Status.Active
			}
			if oldConfig.Status.Active != nil {
				oldConfigActive = *oldConfig.Status.Active
			}
			if !oldConfigActive && newConfigActive {
				return true
			}
			return false

		},
	}
}

// RabbitmqActiveChange returns predicate function based on group kind.
func RabbitmqActiveChange() predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldRabbitmq, ok := e.ObjectOld.(*v1alpha1.Rabbitmq)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			newRabbitmq, ok := e.ObjectNew.(*v1alpha1.Rabbitmq)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			newRabbitmqActive := false
			oldRabbitmqActive := false
			if newRabbitmq.Status.Active != nil {
				newRabbitmqActive = *newRabbitmq.Status.Active
			}
			if oldRabbitmq.Status.Active != nil {
				oldRabbitmqActive = *oldRabbitmq.Status.Active
			}
			if !oldRabbitmqActive && newRabbitmqActive {
				return true
			}
			return false

		},
	}
}

// ZookeeperActiveChange returns predicate function based on group kind.
func ZookeeperActiveChange() predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldZookeeper, ok := e.ObjectOld.(*v1alpha1.Zookeeper)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			newZookeeper, ok := e.ObjectNew.(*v1alpha1.Zookeeper)
			if !ok {
				reqLogger.Info("type conversion mismatch")
			}
			newZookeeperActive := false
			oldZookeeperActive := false
			if newZookeeper.Status.Active != nil {
				newZookeeperActive = *newZookeeper.Status.Active
			}
			if oldZookeeper.Status.Active != nil {
				oldZookeeperActive = *oldZookeeper.Status.Active
			}
			if !oldZookeeperActive && newZookeeperActive {
				return true
			}
			return false

		},
	}
}

// MergeCommonConfiguration combines common configuration of manager and service.
func MergeCommonConfiguration(manager v1alpha1.ManagerConfiguration,
	instance v1alpha1.PodConfiguration) v1alpha1.PodConfiguration {
	if len(instance.NodeSelector) == 0 && len(manager.NodeSelector) > 0 {
		instance.NodeSelector = manager.NodeSelector
	}
	if instance.HostNetwork == nil && manager.HostNetwork != nil {
		instance.HostNetwork = manager.HostNetwork
	}
	if len(instance.ImagePullSecrets) == 0 && len(manager.ImagePullSecrets) > 0 {
		instance.ImagePullSecrets = manager.ImagePullSecrets
	}
	if len(instance.Tolerations) == 0 && len(manager.Tolerations) > 0 {
		instance.Tolerations = manager.Tolerations
	}
	return instance
}

// GetContainerFromList gets a container from a list of container
func GetContainerFromList(containerName string, containerList []*v1alpha1.Container) *v1alpha1.Container {
	for _, instanceContainer := range containerList {
		if containerName == instanceContainer.Name {
			return instanceContainer
		}
	}
	return nil
}

// Check if some labeled pods switch to Running or from Running to another phase
func PodPhaseChanges(podLabels map[string]string) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// TODO Select our pods using labels
			for key, value := range e.MetaOld.GetLabels() {
				if podLabels[key] == value {
					oldPod, ok := e.ObjectOld.(*corev1.Pod)
					if !ok {
						reqLogger.Info("type conversion mismatch")
					}
					newPod, ok := e.ObjectNew.(*corev1.Pod)
					if !ok {
						reqLogger.Info("type conversion mismatch")
					}
					if (newPod.Status.Phase == "Running" && oldPod.Status.Phase != "Running") ||
						(newPod.Status.Phase != "Running" && oldPod.Status.Phase == "Running") {
						return true
					}
				}
			}
			return false
		},
	}
}
