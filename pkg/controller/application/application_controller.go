package application

import (
	"context"
	appsv1alpha1 "github.com/blrn/app-operator/pkg/apis/apps/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Application Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileApplication{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("application-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Application
	err = c.Watch(&source.Kind{Type: &appsv1alpha1.Application{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Application
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.Application{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.Application{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &extensionsv1beta1.Ingress{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.Application{},
	})

	return nil
}

var _ reconcile.Reconciler = &ReconcileApplication{}

// ReconcileApplication reconciles a Application object
type ReconcileApplication struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

type AppResources struct {
	deployment *appsv1.Deployment
	service    *corev1.Service
	ingress    *extensionsv1beta1.Ingress
}

// Reconcile reads that state of the cluster for a Application object and makes changes based on the state read
// and what is in the Application.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileApplication) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling Application %s/%s\n", request.Namespace, request.Name)

	// Fetch the Application instance
	instance := &appsv1alpha1.Application{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	resources := r.newAppResourcesForCR(instance)

	// Check if the deployment already exists
	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: resources.deployment.Name, Namespace: resources.deployment.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Creating a new deployment %s/%s\n", resources.deployment.Namespace, resources.deployment.Name)
		err = r.client.Create(context.TODO(), resources.deployment)
		if err != nil {
			log.Printf("error creating deployment: %+v", err)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	} else {
		log.Printf("deployment already exists")
	}

	if resources.service != nil {
		// Check if the service already exists
		found := &corev1.Service{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: resources.service.Name, Namespace: resources.service.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			log.Printf("Creating a new service %s/%s\n", resources.service.Namespace, resources.service.Name)
			err = r.client.Create(context.TODO(), resources.service)
			if err != nil {
				log.Printf("error creating service: %+v\n", err)
				return reconcile.Result{}, err
			}
		} else if err != nil {
			return reconcile.Result{}, err
		} else {
			log.Printf("sesrvice already exists\n")
		}
	}

	if resources.ingress != nil {
		found := &extensionsv1beta1.Ingress{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: resources.ingress.Name, Namespace: resources.ingress.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			log.Printf("Creating a new ingress %s/%s\n", resources.ingress.Namespace, resources.ingress.Name)
			err = r.client.Create(context.TODO(), resources.ingress)
			if err != nil {
				log.Printf("error creating ingress: %+v\n", err)
				return reconcile.Result{}, err
			}
		} else if err != nil {
			return reconcile.Result{}, err
		} else {
			log.Printf("ingress already exists\n")
		}
	}

	// All resources exists don't requeue
	return reconcile.Result{}, nil
}

// newAppResourcesForCR returns a AppResources object for the cr
func (r *ReconcileApplication) newAppResourcesForCR(cr *appsv1alpha1.Application) (res AppResources) {
	res.deployment = r.newDeploymentForCR(cr)
	res.ingress = r.newIngressForCR(cr)
	res.service = r.newServiceForCR(cr)
	return
}

func labelsForCR(cr *appsv1alpha1.Application) map[string]string {
	return map[string]string{"app": cr.Name}
}

func (r *ReconcileApplication) newIngressForCR(cr *appsv1alpha1.Application) *extensionsv1beta1.Ingress {
	spec := cr.Spec
	if spec.Ingress == nil {
		return nil
	}
	path := spec.Ingress.Path
	if path == "" {
		path = "/"
	}
	targetPort := spec.Ingress.TargetPort
	if spec.Ingress.TargetPort == 0 && spec.Service != nil {
		targetPort = spec.Service.Port
	}
	labels := labelsForCR(cr)

	serviceName := getServiceNameForCR(cr)

	// create a deep copy so we don't actually modify the annotation of the CR
	ingressAnnotations := make(map[string]string)
	for k, v := range cr.Annotations {
		ingressAnnotations[k] = v
	}
	if _, ok := ingressAnnotations["kubernetes.io/ingress.class"]; !ok {
		ingressAnnotations["kubernetes.io/ingress.class"] = "traefik"
	}

	ingSpec := extensionsv1beta1.IngressSpec{}
	ingSpec.Rules = []extensionsv1beta1.IngressRule{
		{
			Host: spec.Ingress.Host,
			IngressRuleValue: extensionsv1beta1.IngressRuleValue{
				HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
					Paths: []extensionsv1beta1.HTTPIngressPath{
						{
							Path: path,
							Backend: extensionsv1beta1.IngressBackend{
								ServiceName: serviceName,
								ServicePort: intstr.IntOrString{IntVal: targetPort},
							},
						},
					},
				},
			},
		},
	}
	ing := &extensionsv1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Name,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: ingressAnnotations,
		},
		Spec: ingSpec,
	}

	controllerutil.SetControllerReference(cr, ing, r.scheme)
	return ing
}

func (r *ReconcileApplication) newServiceForCR(cr *appsv1alpha1.Application) *corev1.Service {
	spec := cr.Spec
	serviceSpec := spec.Service
	if serviceSpec == nil {
		return nil
	}
	name := getServiceNameForCR(cr)
	labels := labelsForCR(cr)
	log.Printf("serviceSpec.Port: %d\n", serviceSpec.Port)
	log.Printf("serviceSpec.TargetPort: %d\n", serviceSpec.TargetPort)
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: cr.Annotations,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Protocol:   "TCP",
					Port:       serviceSpec.Port,
					TargetPort: intstr.IntOrString{Type: 0, IntVal: serviceSpec.TargetPort},
				},
			},
			Selector: labels,
		},
	}
	controllerutil.SetControllerReference(cr, svc, r.scheme)
	return svc
}

func getServiceNameForCR(cr *appsv1alpha1.Application) string {
	name := cr.Spec.Service.Name
	if name == "" {
		name = cr.Name
	}
	return name
}

func (r *ReconcileApplication) newDeploymentForCR(cr *appsv1alpha1.Application) *appsv1.Deployment {
	labels := labelsForCR(cr)
	spec := cr.Spec
	replicas := spec.Replicas
	if replicas == 0 {
		replicas = 1
	}
	container := newContainerForCR(cr)
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{container},
				},
			},
		},
	}
	controllerutil.SetControllerReference(cr, dep, r.scheme)
	return dep
}

func newContainerForCR(cr *appsv1alpha1.Application) corev1.Container {
	spec := cr.Spec
	con := corev1.Container{
		Image:   spec.Image,
		Name:    cr.Name,
		Command: spec.Command,
		Args:    spec.Args,
		Env:     spec.Env,
	}
	// TODO: Ports if service is defined
	return con
}
