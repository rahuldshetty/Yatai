package services

import (
	"context"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai/api-server/models"

	servingv1alpha2 "github.com/bentoml/yatai-deployment/apis/serving/v1alpha2"
)

type kubeBentoDeploymentService struct{}

var KubeBentoDeploymentService = kubeBentoDeploymentService{}

func (s *kubeBentoDeploymentService) Deploy(ctx context.Context, deploymentTarget *models.DeploymentTarget, deployOption *models.DeployOption) (kubeBentoDeployment *servingv1alpha2.BentoDeployment, err error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		err = errors.Wrap(err, "failed to get associated deployment")
		return
	}

	cli, err := DeploymentService.GetKubeBentoDeploymentCli(ctx, deployment)
	if err != nil {
		err = errors.Wrap(err, "failed to get kube bento deployment cli")
		return
	}

	if deploymentTarget.Config != nil && deploymentTarget.Config.KubeResourceVersion != "" {
		var oldKubeBentoDeployment *servingv1alpha2.BentoDeployment
		oldKubeBentoDeployment, err = cli.Get(ctx, deployment.Name, metav1.GetOptions{})
		isNotFound := apierrors.IsNotFound(err)
		if err != nil && !isNotFound {
			err = errors.Wrap(err, "failed to get kube bento deployment")
			return
		}
		if !isNotFound && oldKubeBentoDeployment.ResourceVersion == deploymentTarget.Config.KubeResourceVersion {
			kubeBentoDeployment = oldKubeBentoDeployment
			return
		}
	}

	defer func() {
		if err != nil {
			return
		}
		status := modelschemas.DeploymentStatusDeploying
		_, _ = DeploymentService.UpdateStatus(ctx, deployment, UpdateDeploymentStatusOption{
			Status: &status,
		})
		deployment.Status = status
		ctx_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		go func() {
			defer cancel()
			_, _ = DeploymentService.SyncStatus(ctx_, deployment)
		}()
	}()

	bento, err := BentoService.GetAssociatedBento(ctx, deploymentTarget)
	if err != nil {
		err = errors.Wrap(err, "failed to get associated bento")
		return
	}
	tag, err := BentoService.GetTag(ctx, bento)
	if err != nil {
		err = errors.Wrap(err, "failed to get bento tag")
		return
	}

	var autoscalingSpec *modelschemas.DeploymentTargetHPAConf
	if deploymentTarget.Config != nil {
		autoscalingSpec = deploymentTarget.Config.HPAConf
	}

	envs := make([]modelschemas.LabelItemSchema, 0)
	if deploymentTarget.Config != nil && deploymentTarget.Config.Envs != nil {
		for _, env := range *deploymentTarget.Config.Envs {
			envs = append(envs, *env)
		}
	}

	var resources *modelschemas.DeploymentTargetResources
	if deploymentTarget.Config != nil {
		resources = deploymentTarget.Config.Resources
	}

	var runners []servingv1alpha2.BentoDeploymentRunnerSpec
	if deploymentTarget.Config != nil && deploymentTarget.Config.Runners != nil {
		runners = make([]servingv1alpha2.BentoDeploymentRunnerSpec, 0, len(deploymentTarget.Config.Runners))
		for name, runner := range deploymentTarget.Config.Runners {
			envs_ := make([]modelschemas.LabelItemSchema, 0)
			if runner.Envs != nil {
				for _, env := range *runner.Envs {
					envs_ = append(envs_, *env)
				}
			}
			runners = append(runners, servingv1alpha2.BentoDeploymentRunnerSpec{
				Name:        name,
				Resources:   runner.Resources,
				Autoscaling: runner.HPAConf,
				Envs:        &envs_,
			})
		}
	}

	ingress := servingv1alpha2.BentoDeploymentIngressSpec{}

	if deploymentTarget.Config != nil && deploymentTarget.Config.EnableIngress != nil && *deploymentTarget.Config.EnableIngress {
		ingress.Enabled = true
	}

	kubeBentoDeployment = &servingv1alpha2.BentoDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.Name,
			Namespace: DeploymentService.GetKubeNamespace(deployment),
		},
		Spec: servingv1alpha2.BentoDeploymentSpec{
			BentoTag:    string(tag),
			Autoscaling: autoscalingSpec,
			Envs:        &envs,
			Resources:   resources,
			Runners:     runners,
			Ingress:     ingress,
		},
	}

	var oldKubeBentoDeployment *servingv1alpha2.BentoDeployment
	oldKubeBentoDeployment, err = cli.Get(ctx, kubeBentoDeployment.Name, metav1.GetOptions{})
	isNotFound := apierrors.IsNotFound(err)
	if err != nil && !isNotFound {
		err = errors.Wrap(err, "failed to get kube bento deployment")
		return
	}
	if isNotFound {
		kubeBentoDeployment, err = cli.Create(ctx, kubeBentoDeployment, metav1.CreateOptions{})
		if err != nil {
			err = errors.Wrapf(err, "failed to create kube bento deployment %s", kubeBentoDeployment.Name)
			return
		}
	} else {
		kubeBentoDeployment.SetResourceVersion(oldKubeBentoDeployment.GetResourceVersion())
		kubeBentoDeployment, err = cli.Update(ctx, kubeBentoDeployment, metav1.UpdateOptions{})
		if err != nil {
			err = errors.Wrapf(err, "failed to update kube bento deployment %s", kubeBentoDeployment.Name)
			return
		}
	}
	return
}
